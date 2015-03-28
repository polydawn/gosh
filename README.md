# gosh [![Build Status](https://travis-ci.org/polydawn/gosh.svg)](https://travis-ci.org/polydawn/gosh) [![Doc Status](https://godoc.org/github.com/polydawn/gosh?status.png)](https://godoc.org/github.com/polydawn/gosh)

---

A simple-minded API for exec'ing in Go.

Avoid all the futz with connecting your inputstreams to your outputstreams.
Get error handling by default instead of return codes (you forgot to check them, admit it).
Skip over miles of boilerplate.
Generally be a happier person.

Example:


```go
        // basic commands
        Sh("head")("--bytes=20", "/dev/zero")()

        // making your own shorthand
        aptget := Sh("apt-get")("install")
        aptyes := aptget("-y")
        // now `aptyes("git")` will install git and the `-y` argument means it will proceed automatically
```

Gosh is meant to be usable as a bash or python replacement in rapid-iteration scripting, as well as heavy duty usage as a library in large applications.

Any kind of collection can be used for input and output of strings and bytes -- buffers, strings, channels, readers, writers -- you name it, it'll fly.
Simple applications set up their input and output, call the command, wait for return, and go about their business.
Fancier applications that want to streaming parallel processing can drop in a channel and parallel the day away.

After every call, a new command object is returned.
(Each command object is immutable.)
So, a series of similar commands can be templated and reused!

Error handling is to panic by default if a command has a non-zero exit code.
This is like running bash with with `set -e` (stop on errors -- i.e., what you should literally always be doing).
If you need to handle other exit codes, you can provide a list of okay codes.

No external dependencies.

Inspired by https://github.com/amoffat/sh/ and https://github.com/polydawn/josh/ .

No warranty implied.  Do not place in closed boxes with cats.  May eat your homework.  *Is* probably web scale (but no one knows what that means).  Artisanally crafted with venom and scorn.  Coroutines included.



## Building

`./goad`


