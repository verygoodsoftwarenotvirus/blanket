# tarp [![Build Status](https://travis-ci.org/verygoodsoftwarenotvirus/tarp.svg?branch=master)](https://travis-ci.org/verygoodsoftwarenotvirus/tarp) [![Coverage Status](https://coveralls.io/repos/github/verygoodsoftwarenotvirus/tarp/badge.svg?branch=master)](https://coveralls.io/github/verygoodsoftwarenotvirus/tarp?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/verygoodsoftwarenotvirus/tarp)](https://goreportcard.com/report/github.com/verygoodsoftwarenotvirus/tarp)

`tarp` is a tool that helps you catch functions which don't have direct unit tests in your Go packages.

## Usage

Say, for example, you had the following Go file:

```go
package simple

func A() string {
    return "A"
}

func B() string {
    return "B"
}

func C() string {
    return "C"
}

func wrapper() {
   A()
   B()
   C()
}
```

and you had the following test for that file:

```go
package simple

import (
    "testing"
)

func TestA(t *testing.T) {
    A()
}

func TestC(t *testing.T) {
    C()
}

func TestWrapper(t *testing.T) {
    wrapper()
}
```

Running `go test -cover` on that package yields the following output:

```bash
PASS
coverage: 100.0% of statements
ok      github.com/verygoodsoftwarenotvirus/tarp/example_packages/simple    0.006s
```

However, note that `B` doesn't have a direct test the way that `A` and `C` do. `B` is only "tested" in `TestWrapper`. Because all the functions are called, `-cover` yields a 100% coverage value. If you ever decide that `wrapper` doesn't need to call `B` anymore, and don't delete the function entirely, you'll have a drop in coverage. What `tarp` seeks to do is catch these sorts of things so that package maintainers can decide what the appropriate course of action is. If you're fine with it, that's cool. If you're not cool with it, then you know what needs to have tests added.

I think `tarp` could also be helpful for new developers looking to contribute towards a project. They can run `tarp` on the package and see if there are some functions they could easily add unit tests for, just to get their feet wet in a project.

## Issues

If you've tried tarp on something and found that it didn't accurately handle some code, or panicked, please feel free to [file an issue](https://github.com/verygoodsoftwarenotvirus/tarp/issues/new). Having an example of the code you experienced issues with is pretty crucial, so keep that in mind.