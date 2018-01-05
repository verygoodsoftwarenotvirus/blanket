# delete previous coverage report, if it exists
if [ -f coverage.out ]
then
    rm coverage.out
fi

# run tests
go install
go test -coverprofile=coverage.out
blanket cover --html=coverage.out

# delete the new coverage report, if it exists, so I don't accidentally commit it to the repo somehow
if [ -f coverage.out ]
then
    rm coverage.out
fi