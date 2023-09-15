# chanio

`chanio` is a Go library for treating a channel as `io.Reader`, `io.Writer` and `io.Closer`. The main motivation is to mimic the `os.Pipe()` function in WASM environment.

## About the Name

I have been searching for existing libraries before I start creating my own. The only finding is https://github.com/mitchellh/iochan, which is doing the reverse thing, i.e. treating `io` readers and writers like channels. So I just reverse the naming to express the reversed intent.
