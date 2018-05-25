# Getting Started with Go
## Installation
Go to [Golang.org download page](https://golang.org/dl/) and download `go1.9.2.darwin-amd64.pkg` 
for your Mac OS. It should include an installer and provide instruction on how to install Go step 
by step. When you are done, make sure you run `go` in your terminal. If Go has been successfully 
installed on your machine, you should expect to see the following being printed to your screen.

```text
MacBook: ~Calvin$ go
Go is a tool for managing Go source coode.

Usage:

        go command [arguments]
```

## Project Structure
### Workspace
Go has this concept of a workspace. It's basically a folder where you'd put all your source code 
for all your Go programs. I usually put my workspace in my home directory. I called my workspace 
`Gopher` but feel free to name it whatever you like.

```text
Calvin
        - Applications
        - Desktop
        - Documents
        - etc...
        - Gopher
                - bin
                - pkg
                - src
```

### Go Path
In order for Go to recognize your workspace directory, you must go to your home directory and 
define your environmental variables in your `.bash_profile`.

For example:
```text
cd ~
atom .bash_profile
```

And then insert the following into your bash profile:
```text
# Go paths
export GO=/user/local/go
export GOPATH=/Users/Calvin/Gopher
export PATH=$PATH:$GO/bin:$GOPATH/bin
```

Now, whenever you start a new project, put it into `$GOPATH/src` folder.

For example:
```text
- Gopher
        - bin
        - pkg
        - src
                - go-academy
                        - first_program
                                - hello_world.go
                                - main.go
                                - bye_world.go
```

## Video 01: Hello World in Go
By now you should have your Go installed, your workspace created, and your `$GOPATH` is pointing to 
the workspace. Let's create our first program in Go!

[Hello World in Go](https://youtu.be/5-FFapKA9sM)

## Compiled Language
In case you don't really know what a compiled language is or how language is being compiled from 
source code to machine code. Here is a fun introduction to the concept.

[How do computers read code?](https://www.youtube.com/watch?v=QXjU9qTsYCc)
