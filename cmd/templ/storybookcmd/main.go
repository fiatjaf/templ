package storybookcmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"sync"

	"github.com/a-h/templ"
	"github.com/rs/cors"
)

func Run(args []string) (err error) {
	helloConf := HandlePreview("Hello", Hello,
		TextArg("stringArg", "the name"),
		IntArg("intArg", 0, 0, 100, 5))
	// TODO: Automatically output the configuration into the storybook ./storybook-server/stories directory.
	conf, _ := json.MarshalIndent(helloConf, " ", "")
	fmt.Println(string(conf))
	return http.ListenAndServe("localhost:60606", cors.Default().Handler(http.DefaultServeMux))
}

func HandlePreview(name string, componentConstructor interface{}, args ...Arg) (c *Conf) {
	//TODO: Check that the component constructor is a function that returns a templ.Component.
	//TODO: Get the function name with reflection instead of using the name parameter.
	c = NewConf(name, args...)
	h := NewHandler(name, componentConstructor, args...)
	http.Handle("/storybook_preview/"+url.PathEscape(name), h)
	return
}

func NewHandler(name string, f interface{}, args ...Arg) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		argv := make([]interface{}, len(args))
		q := r.URL.Query()
		for i, arg := range args {
			argv[i] = arg.Get(q)
		}
		component, err := eval(name, f, argv)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		templ.Handler(component).ServeHTTP(w, r)
	})
}

func NewConf(title string, args ...Arg) *Conf {
	c := &Conf{
		Title: title,
		Parameters: StoryParameters{
			Server: map[string]interface{}{
				"id": title,
			},
		},
		Args:     NewSortedMap(),
		ArgTypes: NewSortedMap(),
		Stories:  []Story{},
	}
	for _, arg := range args {
		c.Args.Add(arg.Name, arg.Value)
		c.ArgTypes.Add(arg.Name, map[string]interface{}{
			"control": arg.Control,
		})
	}
	c.AddStory("Default")
	return c
}

func (c *Conf) AddStory(name string, args ...Arg) {
	m := NewSortedMap()
	for _, arg := range args {
		m.Add(arg.Name, arg.Value)
	}
	c.Stories = append(c.Stories, Story{
		Name: name,
		Args: NewSortedMap(),
	})
}

// Controls for the configuration.
// See https://storybook.js.org/docs/react/essentials/controls
type Arg struct {
	Name    string
	Value   interface{}
	Control interface{}
	Get     func(q url.Values) interface{}
}

func TextArg(name, value string) Arg {
	return Arg{
		Name:    name,
		Value:   value,
		Control: "text",
		Get: func(q url.Values) interface{} {
			return q.Get(name)
		},
	}
}

func BooleanArg(name string, value bool) Arg {
	return Arg{
		Name:    name,
		Value:   value,
		Control: "boolean",
		Get: func(q url.Values) interface{} {
			return q.Get(name) == "true"
		},
	}
}

func IntArg(name string, value, min, max, step int) Arg {
	return Arg{
		Name:  name,
		Value: value,
		Control: map[string]interface{}{
			"type": "number",
			"min":  min,
			"max":  max,
			"step": step,
		},
		Get: func(q url.Values) interface{} {
			i, _ := strconv.ParseInt(q.Get(name), 10, 64)
			return int(i)
		},
	}
}

func FloatArg(name string, value float64, min, max, step float64) Arg {
	return Arg{
		Name:  name,
		Value: value,
		Control: map[string]interface{}{
			"type": "number",
			"min":  min,
			"max":  max,
			"step": step,
		},
		Get: func(q url.Values) interface{} {
			i, _ := strconv.ParseFloat(q.Get(name), 64)
			return i
		},
	}
}

type Conf struct {
	Title      string          `json:"title"`
	Parameters StoryParameters `json:"parameters"`
	Args       *SortedMap      `json:"args"`
	ArgTypes   *SortedMap      `json:"argTypes"`
	Stories    []Story         `json:"stories"`
}

type StoryParameters struct {
	Server map[string]interface{} `json:"server"`
}

func NewSortedMap() *SortedMap {
	return &SortedMap{
		m:        new(sync.Mutex),
		internal: map[string]interface{}{},
		keys:     []string{},
	}
}

type SortedMap struct {
	m        *sync.Mutex
	internal map[string]interface{}
	keys     []string
}

func (sm *SortedMap) Add(key string, value interface{}) {
	sm.m.Lock()
	defer sm.m.Unlock()
	sm.keys = append(sm.keys, key)
	sm.internal[key] = value
}

func (sm *SortedMap) MarshalJSON() ([]byte, error) {
	sm.m.Lock()
	defer sm.m.Unlock()
	b := new(bytes.Buffer)
	b.WriteRune('{')
	enc := json.NewEncoder(b)
	for i, k := range sm.keys {
		enc.Encode(k)
		b.WriteRune(':')
		enc.Encode(sm.internal[k])
		if i < len(sm.keys)-1 {
			b.WriteRune(',')
		}
	}
	b.WriteRune('}')
	return b.Bytes(), nil
}

type Story struct {
	Name string `json:"name"`
	Args *SortedMap
}

// Run the template.

var stringType = reflect.TypeOf("").Kind()
var intType = reflect.TypeOf(0).Kind()

func eval(name string, fn interface{}, values []interface{}) (output templ.Component, err error) {
	v := reflect.ValueOf(fn)
	t := v.Type()
	argv := make([]reflect.Value, t.NumIn())
	if len(argv) != len(values) {
		err = fmt.Errorf("storybook templ: component %s expects %d argument, but %d were provided", fn, len(argv), len(values))
		return
	}
	for i := 0; i < len(argv); i++ {
		switch t.In(i).Kind() {
		case stringType:
			argv[i] = reflect.ValueOf(values[i])
		case intType:
			argv[i] = reflect.ValueOf(values[i])
			//TODO: Match other types, or return an error if the type cant be handled.
		}
	}
	result := v.Call(argv)
	if len(result) != 1 {
		err = fmt.Errorf("storybook templ: function %s must return a templ.Component", name)
		return
	}
	output, ok := result[0].Interface().(templ.Component)
	if !ok {
		err = fmt.Errorf("storybook templ: result of function %s is not a templ.Component", name)
		return
	}
	return output, nil
}

// Example template for testing.

func Hello(stringArg string, intArg int) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		io.WriteString(w, "<h1>Hello</h1>")
		io.WriteString(w, "<h1>"+stringArg+"</h1>")
		io.WriteString(w, "<h1>"+fmt.Sprintf("%d", intArg)+"</h1>")
		return nil
	})
}
