# get-google-tokens

## How to use
### install
```bash
git clone https://github.com/abekoh/get-google-tokens.git
cd get-google-tokens
go install
```
### run
```bash
get-google-tokens -json client_secret_XXX.apps.googleusercontent.com.json -scope https://www.googleapis.com/auth/photoslibrary.appendonly
```
After running, access shown URL and authorize.
You can get access token and refresh token.
