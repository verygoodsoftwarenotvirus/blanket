# tarp [![Build Status](https://travis-ci.org/verygoodsoftwarenotvirus/tarp.svg?branch=master)](https://travis-ci.org/verygoodsoftwarenotvirus/tarp) [![Coverage Status](https://coveralls.io/repos/github/verygoodsoftwarenotvirus/tarp/badge.svg?branch=master)](https://coveralls.io/github/verygoodsoftwarenotvirus/tarp?branch=master)

`tarp` is a tiny helper library that helps you catch functions which don't have direct unit tests in your Go packages

So say, for instance, you had the following Go file:

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

Note that `B` doesn't have a direct test the way that `A` and `C` do. `B` is, however, tested in `TestWrapper`. This would result in a 100% coverage report, because all the functions are called. What `tarp` seeks to do is catch these sorts of things so that package maintainers can decide what the appropriate course of action is. If you're cool with it, that's cool. If you're not cool with it, then you know what needs to have tests added.

I think `tarp` could also be helpful for new developers looking to contribute towards a project. They can run `tarp` on the package and see if there are some functions they could easily add unit tests for, just to get their feet wet in a project.

`tarp` is currently in active development, and is not yet suitable for use or release. I will update this README when I feel this is no longer the case.

## Issues

If you've tried tarp on something and found that it didn't accurately handle some code, please feel free to [file an issue](https://github.com/verygoodsoftwarenotvirus/tarp/issues/new). Having an example of the code you experienced issues with is pretty crucial, so keep that in mind.