# delete previous coverage report, if it exists
if [ -f coverage.out ]
then
    rm coverage.out
fi

# run tests
go test -coverpkg=github.com/verygoodsoftwarenotvirus/tarp -coverprofile=tarp.cover.out
go test -coverpkg=github.com/verygoodsoftwarenotvirus/tarp/cmd -coverprofile=tarp.cmd.cover.out

# big thanks to https://lk4d4.darth.io/posts/multicover/ for this, it's gross, but it's not their (or my) fault
echo "mode: set" > coverage.out && cat *.cover.out | grep -v mode: | sort -r | \
awk '{if($1 != last) {print $0;last=$1}}' >> coverage.out

go tool cover -html=coverage.out

rm *.cover.out
# delete the new coverage report, if it exists, so I don't accidentally commit it to the repo somehow
if [ -f coverage.out ]
then
    rm coverage.out
fi