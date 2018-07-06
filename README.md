# perpetual [![Build Status](https://travis-ci.org/brandur/perpetual.svg?branch=master)](https://travis-ci.org/brandur/perpetual)

An experiment in long-term thinking.

``` sh
go get -u github.com/aws/aws-lambda-go/lambda
go get -u github.com/dghubble/oauth1
go get -u github.com/golang/lint/golint
go get -u github.com/stretchr/testify/require

make
```

## Setting interval messages

``` sh
mv intervals.go.sample intervals.go
# edit intervals.go
```

## Getting an access token

After creating an app, Twitter allows you to create a
person access token and access token secret pair for it.
This will be under a section called _Your Access Token_ on
the app's keys page.

You will need to copy out all four of your consumer key,
secret, access token, and access token secret.

## Lambda

1. Use `make package` to create a `.zip` to upload.
2. Set "Handler" (under "Function Code") to the name of the
   zip file, `perpetual`.
3. Set environmental variables for each of the four keys
   above.
4. Set a tag for `app=perpetual` to make these easy to
   find.
