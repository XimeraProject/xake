# /bin/bash

echo Bump version...
sed -ri 's/(.*)(app\.Version = )(\"[0-9]+\.[0-9]+\.)([0-9]+)(.*)/echo "\1\2\\"\3$((\4+1))\\"\5"/ge' main.go

TAG=v$(cat main.go | egrep '^[[:space:]]+app.Version = "[0-9]+\.[0-9]+\.[0-9]+"'  | sed 's/.*=//g' | tr -d '" ')
echo Tagging this commit as $TAG
git add main.go
git commit -m "Bumping version to $TAG"
echo Creating and pushing tag
git push
git tag -a $TAG -m "Version $TAG" && git push origin $TAG
