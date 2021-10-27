# friendly-captcha-go-sdk
A Go client for the Friendly Captcha verification API.

## Installation
```
go get github.com/friendlycaptcha/friendly-captcha-go-sdk
```

## Usage

```go
import "github.com/friendlycaptcha/friendly-captcha-go-sdk"

frcClient := friendlycaptcha.NewClient(apikey, sitekey)
```


```go
// In your middleware or request handler
solution := r.FormValue(friendlycaptcha.SolutionFormFieldName)
shouldAccept, err := frcClient.CheckCaptchaSolution(r.Context(), solution)

if err != nil {
    if errors.Is(err, friendlycaptcha.ErrVerificationFailedDueToClientError) {
        log.Printf("!!!!!\n\nFriendlyCaptcha is misconfigured! Check your Friendly Captcha API key and sitekey: %v\n", err)
        // Send yourself an alert - the captcha won't be able to do its job to prevent spam.
    } else if (errors.Is(err, friendlycaptcha.ErrVerificationRequest)) {
        log.Printf("Could not talk to the Friendly Captcha API: %v\n", err)
        // Perhaps the Friendly Captcha API is down?
    }
}

if !shouldAccept { // The captcha was invalid
    // Show the user a message that the anti-robot verification failed and that they should try again
    return
}

// The captcha check was succesful, handle the request :)
```

Beware that the `CheckCaptchaSolution` function returns two values:
 * Whether you should accept the request (`bool`)
 * An error (or nil)
 
 Even if the error is non-nil, the first boolean value may still be true and you should accept the request! 
### Advanced, optional strictness setting
As a best practice we accept the captcha solution if we are unable to verify it: if we misconfigure our apikey or Friendly Captcha's API goes down we would rather accept all requests than lock all users out. 

 If you want to change this behavior you can set `client.Strict` to true, then the accept value will only be true if we were actually able to verify the captcha solution and it was valid.


## Example

Run the example
```shell
cd examples/form
FRC_SITEKEY=<my sitekey> FRC_APIKEY=<my api key> go run main.go
```

Then open your browser and head to [http://localhost:8844](http://localhost:8844)

> Note: you can create a sitekey and API key in the [Friendly Captcha dashboard](https://app.friendlycaptcha.com/account).

**Example Screenshot**

![Example screenshot](https://i.imgur.com/bsp7qDA.png)

## License
[MIT](./LICENSE)