# openrvs-registry Developer Documentation

This repo uses [Go](https://golang.org/), a modern, cross-platform, concurrent, compiled, garbage-collected, statically-typed programming language with an [extensive standard library](https://golang.org/pkg/#stdlib).

## Initial Setup

1. Download and install the Go programming language for your OS here: https://golang.org/doc/install
1. Create a directory to contain all Go code, such as `%USERPROFILE%\go` (recommended)
1. Make sure the environment variable `GOPATH` is set to the above directory
1. Try to download openrvs-registry with `go get github.com/ijemafe/openrvs-registry` on the command line. Go has tight coupling with Git, and this clones the repository under the hood

## Building the Code

Assuming a Windows development environment, there is a batch file to generate builds for both 64-bit Windows and 64-bit Linux at the same time:

```bash
> cd %GOPATH%\src\github.com\ijemafe\openrvs-registry\cmd\registry
> build.bat
```

The Windows build is `registry.exe`, and the Linux build is simply `registry`.

## Running the Code

Run `registry.exe` to run the build locally. All log information is printed to `stdout` and displayed in the terminal window:

```bash
> registry.exe
2020/05/30 23:35:27 openrvs-registry process started
2020/05/30 23:35:27 loading servers from file
2020/05/30 23:35:27 reading checkpoint file at checkpoint.csv
2020/05/30 23:35:27 there are now 48 registered servers (confirm over http)
2020/05/30 23:35:27 starting http listener
2020/05/30 23:35:27 starting udp listener
```

You can now hit the HTTP URLs in your browser at `http://localhost:8080/<path>`,
or send UDP beacons to `udp://localhost:8080` to test automatic registration.

If you want to run the app from a different working directory, you can:

```bash
> registry.exe -csvdir=C:\path\to\csv\files\\
```

The trailing slash must be included, and on Windows there must be two (since `\` is typically an escape character). On Linux, use forward slashes instead.

If you want to run locally without compiling a new build, you can:

```bash
> cd %GOPATH%\src\github.com\ijemafe\openrvs-registry\cmd\registry
> go run main.go
```

Now you can tweak the code and repeat either set of steps above to iterate on changes.

## Editing Code

I recommend [VSCode](https://code.visualstudio.com/) from Microsoft for writing Go code on Windows. It's free, and when you open a `.go` file for the first time, it will automatically prompt you to install the Go extension.

The most useful buttons are in the top-left. From top to bottom: "Explorer" for organizing files in a repo, "Search" for finding strings across all files, and "Source Control" for the built-in Git integration. You can create branches, commit, push, and pull from inside VSCode.

In the `File` menu, click `Open Folder` and select `openrvs-registry`. This will automatically detect the Git repository.

## Navigating the Code

Currently, there are five files containing Go code:

1. `cmd/registry/main.go`: the primary code. starts the http and udp listeners,
	schedules disk checkpointing, and schedules healthchecks.
1. `csv.go`: contains code for converting CSV to Server objects and vice versa
1. `healthcheck.go`: contains logic for hiding unhealthy servers
1. `latest.go`: contains code for hitting the Github API
1. `types.go`: contains definitions and utility code unlikely to change

## Logging Errors

Whenever a function returns an `error`, you should log it:

```go
// assume this function returns type 'error' as the second return value
result, err := doSomething()
// 'nil' is the same as none in other languages
if err != nil {
	// the 'log' package writes timestamped logs to stdout
	// use 'fmt' package for logs without timestamps
	log.Println("there was an error:", err)
}
```

## Deep Dive: Threads and Concurrency

Go has built-in support for running code [concurrently](https://en.wikipedia.org/wiki/Concurrency_\(computer_science\)).

To run code in a concurrent thread (called a "goroutine"), use the `go` keyword:

```go
// define a simple function
func sayHello() { fmt.Println("Hello") }
// run that function in its own thread
go sayHello()
```

The above code will now run in a concurrent thread. However, it's not doing
enough to demonstrate the utility of concurrency. Let's say we want to start a
server, and then send some requests to it after some time has passed:

```go
// start a server in a new thread
go startServer()
// start a second thread that sends requests to the server after waiting
// uses an anonymous function of the form 'go func() { ... }()'
go func() {
	time.Sleep(5 * time.Second)
	sendRequestsToServer()
}()
// this code runs as soon as the second thread is fired off
doSomethingElse()
```

As a more complex example, imagine we want to ping several servers and get the
results. Assume it takes 3 seconds to ping each server, and that there are 10
servers. If executed serially, this would take 30 seconds. By using concurrency
to parallelize the work, we can get this back down to 3 seconds regardless of
the number of servers:

```go
// The 'sync' package in the standard library provides helpers for synchronizing
// goroutines. WaitGroup is used to keep track of several concurrent threads.
var wg sync.WaitGroup
var servers []Server
for i, s := range servers {
	// For each server, add a unit of work to the WaitGroup.
	wg.Add(1)
	go func() {
		// Ping each server in its own thread.
		s.Ping()
		// Now that the work is done, remove the unit of work.
		wg.Done()
	}()
}
// When all units of work are complete, allow the code to move on.
wg.Wait()
analyzeResults()
```

When two threads attempt to access the same memory, the error can cause a panic,
crashing the app. For this reason, there are some safety features we can use to
let the scheduler queue our memory access for us.

Consider these two programs:

```go
// 'make' allocates an object in memory to prepare it for population.
// 'example' maps string keys to integer values.
example := make(map[string]int)
go func() {
	example["foo"] = 1
}()
go func() {
	example["foo"] = 2
}()
```

```go
example := make(map[string]int)
// Mutex = mutual exclusion: access by one prevents access by others
// RWMutex from the 'sync' package prevents both reads and writes when locked.
lock := sync.RWMutex{}
go func() {
	lock.Lock()
	example["foo"] = 1
	lock.Unlock()
}()
go func() {
	lock.Lock()
	example["foo"] = 2
	lock.Unlock()
}()
```

The first program can cause a panic and crash, because two threads attempt to
simultaneously write to the same memory. In the second program, calls to
`lock.Lock()` will block until the lock holder calls `lock.Unlock()`, allowing
the scheduler to queue access until the lock is released. This prevents the
simultaneous access and the panic!
