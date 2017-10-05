package main

import (
	"bytes"
	"fmt"
	"go/token"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"golang.org/x/tools/cover"

	"github.com/bouk/monkey"
	"github.com/fatih/set"
	"github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////
//                                                    //
//               Test Helper Functions                //
//                                                    //
////////////////////////////////////////////////////////

func buildSelfPath(t *testing.T) string {
	t.Helper()
	gopath := os.Getenv("GOPATH")
	return strings.Join([]string{gopath, "src", "github.com", "verygoodsoftwarenotvirus", "tarp"}, "/")
}

////////////////////////////////////////////////////////
//                                                    //
//                   Actual Tests                     //
//                                                    //
////////////////////////////////////////////////////////

func TestFindFile(t *testing.T) {
	shouldFail := func(t *testing.T) {
		_, err := findFile("test")
		assert.NotNil(t, err)
	}
	t.Run("should fail", shouldFail)

	shouldSucceed := func(t *testing.T) {
		_, err := findFile("cmd/internal/objfile")
		assert.Nil(t, err)
	}
	t.Run("should succeed", shouldSucceed)
}

func TestHTMLOutput(t *testing.T) {
	simpleMainPath := fmt.Sprintf("%s/main.go", buildExamplePackagePath(t, "simple", true))
	exampleProfilePath := "example_files/simple_coverage.out"
	exampleReport := tarpReport{
		Called:   set.New("a", "c", "wrapper"),
		Declared: set.New("a", "b", "c", "wrapper"),
		DeclaredDetails: map[string]tarpFunc{
			"a": {
				Name:     "a",
				Filename: simpleMainPath,
				DeclPos: token.Position{
					Filename: simpleMainPath,
					Offset:   16,
					Line:     3,
					Column:   1,
				},
				RBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   32,
					Line:     3,
					Column:   17,
				},
				LBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   46,
					Line:     5,
					Column:   1,
				},
			},
			"b": {
				Name:     "b",
				Filename: simpleMainPath,
				DeclPos: token.Position{
					Filename: simpleMainPath,
					Offset:   49,
					Line:     7,
					Column:   1,
				},
				RBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   65,
					Line:     7,
					Column:   17,
				},
				LBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   79,
					Line:     9,
					Column:   1,
				},
			},
			"c": {
				Name:     "c",
				Filename: simpleMainPath,
				DeclPos: token.Position{
					Filename: simpleMainPath,
					Offset:   82,
					Line:     11,
					Column:   1,
				},
				RBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   98,
					Line:     11,
					Column:   17,
				},
				LBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   112,
					Line:     13,
					Column:   1,
				},
			},
			"wrapper": {
				Name:     "wrapper",
				Filename: simpleMainPath,
				DeclPos: token.Position{
					Filename: simpleMainPath,
					Offset:   115,
					Line:     15,
					Column:   1,
				},
				RBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   130,
					Line:     15,
					Column:   16,
				},
				LBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   147,
					Line:     19,
					Column:   1,
				},
			},
		},
	}

	tmpFile := fmt.Sprintf("%s/temp.html", buildSelfPath(t))

	err := htmlOutput(exampleProfilePath, tmpFile, exampleReport)
	if err != nil {
		log.Println("htmlOutput should not return an error")
		t.FailNow()
	}

	expected := `
<!DOCTYPE html>
<html>
	<head>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
		<style>
			body {
				background: black;
				color: rgb(80, 80, 80);
			}
			body, pre, #legend span {
				font-family: Menlo, monospace;
				font-weight: bold;
			}
			#topbar {
				background: black;
				position: fixed;
				top: 0; left: 0; right: 0;
				height: 42px;
				border-bottom: 1px solid rgb(80, 80, 80);
			}
			#content {
				margin-top: 50px;
			}
			#nav, #legend {
				float: left;
				margin-left: 10px;
			}
			#legend {
				margin-top: 12px;
			}
			#nav {
				margin-top: 10px;
			}
			#legend span {
				margin: 0 1px;
			}
			.cov0 { color: rgb(192, 0, 0) }
			.cov1 { color: rgb(128, 128, 128) }
			.cov2 { color: rgb(116, 140, 131) }
			.cov3 { color: rgb(104, 152, 134) }
			.cov4 { color: rgb(92, 164, 137) }
			.cov5 { color: rgb(80, 176, 140) }
			.cov6 { color: rgb(68, 188, 143) }
			.cov7 { color: rgb(56, 200, 146) }
			.cov8 { color: rgb(44, 212, 149) }
			.cov9 { color: rgb(32, 224, 152) }
			.cov10 { color: rgb(20, 236, 155) }
			.tarp-uncovered { color: rgb(252, 242, 106) }

		</style>
	</head>
	<body>
		<div id="topbar">
			<div id="nav">
				<select id="files">

				<option value="file0">github.com/verygoodsoftwarenotvirus/tarp/example_packages/simple/main.go (100.0%)</option>

				</select>
			</div>
			<div id="legend">
				<span>not tracked</span>

				<span class="cov0">no coverage</span>
				<span class="cov1">low coverage</span>
				<span class="cov2">*</span>
				<span class="cov3">*</span>
				<span class="cov4">*</span>
				<span class="cov5">*</span>
				<span class="cov6">*</span>
				<span class="cov7">*</span>
				<span class="cov8">*</span>
				<span class="cov9">*</span>
				<span class="cov10">high coverage</span>

				<span class="tarp-uncovered">indirectly tested</span>
			</div>
		</div>
		<div id="content">

		<pre class="file" id="file0" >package simple

func a() string <span class="cov10" title="2">{
    return "A"
}</span>

func b() string <span class="tarp-uncovered" title="1">{
    return "B"
}</span>

func c() string <span class="cov10" title="2">{
    return "C"
}</span>

func wrapper() <span class="cov1" title="1">{
    a()
    b()
    c()
}</span>
</pre>

		</div>
	</body>
	<script>
	(function() {
		var files = document.getElementById('files');
		var visible = document.getElementById('file0');
		files.addEventListener('change', onChange, false);
		function onChange() {
			visible.style.display = 'none';
			visible = document.getElementById(files.value);
			visible.style.display = 'block';
			window.scrollTo(0, 0);
		}
	})();
	</script>
</html>
`

	f, err := ioutil.ReadFile(tmpFile)
	assert.Nil(t, err)

	actual := string(f)
	assert.Equal(t, expected, actual, "output should match expectation")

	err = os.Remove(tmpFile)
	if err != nil {
		log.Printf(`Unable to delete file "%s", be sure to delete it.`, tmpFile)
	}
}

func TestHTMLGen(t *testing.T) {
	simpleMainPath := fmt.Sprintf("%s/main.go", buildExamplePackagePath(t, "simple", true))
	exampleReport := tarpReport{
		Called:   set.New("a", "c", "wrapper"),
		Declared: set.New("a", "b", "c", "wrapper"),
		DeclaredDetails: map[string]tarpFunc{
			"a": {
				Name:     "a",
				Filename: simpleMainPath,
				DeclPos: token.Position{
					Filename: simpleMainPath,
					Offset:   16,
					Line:     3,
					Column:   1,
				},
				RBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   32,
					Line:     3,
					Column:   17,
				},
				LBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   46,
					Line:     5,
					Column:   1,
				},
			},
			"b": {
				Name:     "b",
				Filename: simpleMainPath,
				DeclPos: token.Position{
					Filename: simpleMainPath,
					Offset:   49,
					Line:     7,
					Column:   1,
				},
				RBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   65,
					Line:     7,
					Column:   17,
				},
				LBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   79,
					Line:     9,
					Column:   1,
				},
			},
			"c": {
				Name:     "c",
				Filename: simpleMainPath,
				DeclPos: token.Position{
					Filename: simpleMainPath,
					Offset:   82,
					Line:     11,
					Column:   1,
				},
				RBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   98,
					Line:     11,
					Column:   17,
				},
				LBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   112,
					Line:     13,
					Column:   1,
				},
			},
			"wrapper": {
				Name:     "wrapper",
				Filename: simpleMainPath,
				DeclPos: token.Position{
					Filename: simpleMainPath,
					Offset:   115,
					Line:     15,
					Column:   1,
				},
				RBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   130,
					Line:     15,
					Column:   16,
				},
				LBracePos: token.Position{
					Filename: simpleMainPath,
					Offset:   147,
					Line:     19,
					Column:   1,
				},
			},
		},
	}

	exampleProfilePath := "example_files/simple_coverage.out"
	profiles, err := cover.ParseProfiles(exampleProfilePath)
	if err != nil {
		log.Printf("error reading profile: %s\n", simpleMainPath)
		t.FailNow()
	}

	src, err := ioutil.ReadFile(simpleMainPath)
	if err != nil {
		log.Printf("error reading file: %s\n", simpleMainPath)
		t.FailNow()
	}

	var buf bytes.Buffer
	err = htmlGen(&buf, src, simpleMainPath, profiles[0].Boundaries(src), exampleReport)
	assert.Nil(t, err)

	expected := `package simple

func a() string <span class="cov10" title="2">{
    return "A"
}</span>

func b() string <span class="tarp-uncovered" title="1">{
    return "B"
}</span>

func c() string <span class="cov10" title="2">{
    return "C"
}</span>

func wrapper() <span class="cov1" title="1">{
    a()
    b()
    c()
}</span>
`
	actual := buf.String()

	assert.Equal(t, expected, actual, "output should match expectation")
}

func TestPercentCovered(t *testing.T) {
	shouldReturnZero := func(t *testing.T) {
		exampleInput := &cover.Profile{}

		expected := 0.0
		actual := percentCovered(exampleInput)

		assert.Equal(t, expected, actual, "percentCovered should return expected output")
	}
	t.Run("should return zero", shouldReturnZero)

	shouldReturn100 := func(t *testing.T) {
		exampleInput := &cover.Profile{
			Blocks: []cover.ProfileBlock{
				{
					Count:   1,
					NumStmt: 1,
				},
			},
		}

		expected := 100.0
		actual := percentCovered(exampleInput)

		assert.Equal(t, expected, actual, "percentCovered should return expected output")
	}
	t.Run("should return one hundred", shouldReturn100)
}

func TestGoose(t *testing.T) {
	assert.Equal(t, runtime.GOOS, goose(), "goose should return runtime.GOOS")
}

func TestStartBrowser(t *testing.T) {
	testURL := "test"

	execCommandCalled := false
	monkey.Patch(goose, func() string { return "darwin" })
	fakeCommand := exec.Command(``, ``)
	monkey.Patch(exec.Command, func(name string, args ...string) *exec.Cmd {
		assert.Equal(t, name, "open", "expected and actual command names should match")
		execCommandCalled = true
		return fakeCommand
	})

	startBrowser(testURL)
	assert.True(t, execCommandCalled)
	monkey.Unpatch(goose)
	monkey.Unpatch(exec.Command)
}

func TestRGB(t *testing.T) {
	withZero := func(t *testing.T) {
		expected := "rgb(192, 0, 0)"
		actual := rgb(0)
		assert.Equal(t, expected, actual, "RGB should return expected output when passed zero as an argument")
	}
	t.Run("with zero", withZero)

	withNoneZero := func(t *testing.T) {
		expected := "rgb(128, 128, 128)"
		actual := rgb(1)
		assert.Equal(t, expected, actual, "RGB should return expected output when passed a number greater than zero as an argument")
	}
	t.Run("with > zero", withNoneZero)
}

func TestCSSColors(t *testing.T) {
	expected := template.CSS(".cov0 { color: rgb(192, 0, 0) }\n\t\t\t.cov1 { color: rgb(128, 128, 128) }\n\t\t\t.cov2 { color: rgb(116, 140, 131) }\n\t\t\t.cov3 { color: rgb(104, 152, 134) }\n\t\t\t.cov4 { color: rgb(92, 164, 137) }\n\t\t\t.cov5 { color: rgb(80, 176, 140) }\n\t\t\t.cov6 { color: rgb(68, 188, 143) }\n\t\t\t.cov7 { color: rgb(56, 200, 146) }\n\t\t\t.cov8 { color: rgb(44, 212, 149) }\n\t\t\t.cov9 { color: rgb(32, 224, 152) }\n\t\t\t.cov10 { color: rgb(20, 236, 155) }\n\t\t\t.tarp-uncovered { color: rgb(252, 242, 106) }\n")
	actual := CSScolors()
	assert.Equal(t, expected, actual, "CSSColors should return expected output")
}
