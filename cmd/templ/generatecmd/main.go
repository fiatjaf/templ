package generatecmd

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"go/format"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	_ "net/http/pprof"

	"github.com/a-h/templ/cmd/templ/generatecmd/proxy"
	"github.com/a-h/templ/cmd/templ/generatecmd/run"
	"github.com/a-h/templ/cmd/templ/visualize"
	"github.com/a-h/templ/generator"
	"github.com/a-h/templ/parser/v2"
	"github.com/cenkalti/backoff/v4"
	"github.com/cli/browser"
	"github.com/rjeczalik/notify"
)

type Arguments struct {
	FileName                        string
	Path                            string
	Watch                           bool
	Command                         string
	ProxyPort                       int
	Proxy                           string
	WorkerCount                     int
	GenerateSourceMapVisualisations bool
	// PPROFPort is the port to run the pprof server on.
	PPROFPort int
}

var defaultWorkerCount = runtime.NumCPU()

func Run(args Arguments) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()
	if args.PPROFPort > 0 {
		go func() {
			_ = http.ListenAndServe(fmt.Sprintf("localhost:%d", args.PPROFPort), nil)
		}()
	}
	go func() {
		select {
		case <-signalChan: // First signal, cancel context.
			fmt.Println("\nCancelling...")
			cancel()
		case <-ctx.Done():
		}
		<-signalChan // Second signal, hard exit.
		os.Exit(2)
	}()
	err = runCmd(ctx, args)
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

func runCmd(ctx context.Context, args Arguments) (err error) {
	start := time.Now()
	if args.Watch && args.FileName != "" {
		return fmt.Errorf("cannot watch a single file, remove the -f or -watch flag")
	}
	if args.FileName != "" {
		return processSingleFile(ctx, args.FileName, args.GenerateSourceMapVisualisations)
	}
	var target *url.URL
	if args.Proxy != "" {
		target, err = url.Parse(args.Proxy)
		if err != nil {
			return fmt.Errorf("failed to parse proxy URL: %w", err)
		}
	}
	if args.ProxyPort == 0 {
		args.ProxyPort = 7331
	}

	if args.WorkerCount == 0 {
		args.WorkerCount = defaultWorkerCount
	}
	if !path.IsAbs(args.Path) {
		args.Path, err = filepath.Abs(args.Path)
		if err != nil {
			return
		}
	}

	var p *proxy.Handler
	if args.Proxy != "" {
		p = proxy.New(args.ProxyPort, target)
	}

	reloadEvents := make(chan int64, 64)

	// Start workers.
	processEvents := make(chan string)
	var startedProcessing, finishedProcessing int64
	for i := 0; i < args.WorkerCount; i++ {
		go func() {
			for path := range processEvents {
				if err := processSingleFile(ctx, path, args.GenerateSourceMapVisualisations); err != nil {
					fmt.Printf("Error processing file: %v\n", err)
				}
				reloadEvents <- atomic.AddInt64(&finishedProcessing, 1)
			}
		}()
	}

	// Start watching.
	notificationEvents := make(chan notify.EventInfo, 128)
	if args.Watch {
		fmt.Println("Watching path:", args.Path)
		if err := notify.Watch(filepath.Join(args.Path, "..."), notificationEvents, notify.Remove, notify.Write); err != nil {
			fmt.Println("Failed to watch path:", args.Path, err)
		}
		notify.Stop(notificationEvents)
	} else {
		fmt.Println("Processing path:", args.Path)
		go func() {
			if err := processChanges(ctx, args.Path, &startedProcessing, processEvents); err != nil {
				fmt.Println("Failed to process changes:", err)
			}
		}()
	}

	var lastReload int64
loop:
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Done")
			close(processEvents)
			break loop
		case event := <-notificationEvents:
			fmt.Println("Notification")
			path := event.Path()
			if !strings.HasSuffix(path, ".templ") {
				continue
			}
			if event.Event() == notify.Remove {
				goFileName := strings.TrimSuffix(path, ".templ") + "_templ.go"
				if err := os.Remove(goFileName); err != nil {
					fmt.Printf("Error removing file: %v\n", err)
				}
				continue
			}
			atomic.AddInt64(&startedProcessing, 1)
			processEvents <- path
		case finished := <-reloadEvents:
			fmt.Println("Reload")
			started := atomic.LoadInt64(&startedProcessing)
			if finished != started {
				continue
			}
			changesFound := finished - lastReload
			isFirstReload := lastReload == 0
			lastReload = finished

			fmt.Printf("Generated code for %d templates in %s\n", changesFound, time.Since(start))
			if args.Command != "" {
				fmt.Printf("Executing command: %s\n", args.Command)
				if _, err := run.Run(ctx, args.Path, args.Command); err != nil {
					fmt.Printf("Error starting command: %v\n", err)
				}
				// Send server-sent event.
				if p != nil {
					p.SendSSE("message", "reload")
				}
			}
			if isFirstReload && p != nil {
				go func() {
					fmt.Printf("Proxying from %s to target: %s\n", p.URL, p.Target.String())
					if err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", args.ProxyPort), p); err != nil {
						fmt.Printf("Error starting proxy: %v\n", err)
					}
				}()
				go func() {
					fmt.Printf("Opening URL: %s\n", p.Target.String())
					if err := openURL(p.URL); err != nil {
						fmt.Printf("Error opening URL: %v\n", err)
					}
				}()
			}
			start = time.Now()
		}
	}

	return nil
}

func shouldSkipDir(dir string) bool {
	if dir == "." {
		return false
	}
	if dir == "vendor" || dir == "node_modules" {
		return true
	}
	_, name := path.Split(dir)
	// These directories are ignored by the Go tool.
	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
		return true
	}
	return false
}

func processChanges(ctx context.Context, path string, startedProcessing *int64, target chan string) (err error) {
	return filepath.WalkDir(path, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err = ctx.Err(); err != nil {
			return err
		}
		if info.IsDir() && shouldSkipDir(path) {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".templ") {
			atomic.AddInt64(startedProcessing, 1)
			target <- path
		}
		return nil
	})
}

func openURL(url string) error {
	backoff := backoff.NewExponentialBackOff()
	var client http.Client
	client.Timeout = 1 * time.Second
	for {
		if _, err := client.Get(url); err == nil {
			break
		}
		d := backoff.NextBackOff()
		log.Printf("Server not ready. Retrying in %v...", d)
		time.Sleep(d)
	}
	return browser.OpenURL(url)
}

func processSingleFile(ctx context.Context, fileName string, generateSourceMapVisualisations bool) error {
	start := time.Now()
	err := compile(ctx, fileName, generateSourceMapVisualisations)
	if err != nil {
		return err
	}
	fmt.Printf("Generated code for %q in %s\n", fileName, time.Since(start))
	return err
}

func compile(ctx context.Context, fileName string, generateSourceMapVisualisations bool) (err error) {
	if err = ctx.Err(); err != nil {
		return
	}

	t, err := parser.Parse(fileName)
	if err != nil {
		return fmt.Errorf("%s parsing error: %w", fileName, err)
	}
	targetFileName := strings.TrimSuffix(fileName, ".templ") + "_templ.go"

	var b bytes.Buffer
	sourceMap, err := generator.Generate(t, &b)
	if err != nil {
		return fmt.Errorf("%s generation error: %w", fileName, err)
	}

	data, err := format.Source(b.Bytes())
	if err != nil {
		return fmt.Errorf("%s source formatting error: %w", fileName, err)
	}

	if err = os.WriteFile(targetFileName, data, 0644); err != nil {
		return fmt.Errorf("%s write file error: %w", targetFileName, err)
	}

	if generateSourceMapVisualisations {
		err = generateSourceMapVisualisation(ctx, fileName, targetFileName, sourceMap)
	}
	return
}

func generateSourceMapVisualisation(ctx context.Context, templFileName, goFileName string, sourceMap *parser.SourceMap) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	var templContents, goContents []byte
	var templErr, goErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		templContents, templErr = os.ReadFile(templFileName)
	}()
	go func() {
		defer wg.Done()
		goContents, goErr = os.ReadFile(goFileName)
	}()
	wg.Wait()
	if templErr != nil {
		return templErr
	}
	if goErr != nil {
		return templErr
	}

	targetFileName := strings.TrimSuffix(templFileName, ".templ") + "_templ_sourcemap.html"
	w, err := os.Create(targetFileName)
	if err != nil {
		return fmt.Errorf("%s sourcemap visualisation error: %w", templFileName, err)
	}
	defer w.Close()
	b := bufio.NewWriter(w)
	defer b.Flush()

	return visualize.HTML(templFileName, string(templContents), string(goContents), sourceMap).Render(ctx, b)
}
