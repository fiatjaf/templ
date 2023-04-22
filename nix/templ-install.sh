# Create the standard environment.
source $stdenv/setup
# Go requires a cache directory so let's point it at one.
mkdir -p /tmp/go-cache
export GOCACHE=/tmp/go-cache
export GOMODCACHE=/tmp/go-cache
# Build the source code.
cd $src/cmd/templ
ls
# Build the templ binary and output it.
mkdir -p $out/bin
go build -o $out/bin/templ
