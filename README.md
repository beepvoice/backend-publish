# backend-publish

Beep backend accepts PUT requests and publishes a protobuf-ed version to a [NATS](htts://nats.io) queue, like some sort of weird HTTP/NATS converter. Also handles authentication of said HTTP requests. Needless to say, relies on a NATS instance being up.

**To run this service securely means to run it behind traefik forwarding auth to `backend-auth`**

## Quickstart

```
go build && ./backend-publish
```

## Environment Variables

Supply environment variables by either exporting them or editing ```.env```.

| ENV | Description | Default |
| ---- | ----------- | ------- |
| LISTEN | Host and port number to listen on | :8080 |
| NATS | Host and port of nats | nats://localhost:4222 |
| SECRET | JWT secret | secret |

## API

All requests need to be passed through `traefik` calling `backend-auth` as Forward Authentication. Otherwise, populate `X-User-Claim` with:

```json
{
  "userid": "<userid>",
  "clientid": "<clientid"
}
```

### Put Bite

```
PUT /conversation/:key/start/:start
```

TODO: Description of what this does cos honestly I have no idea Ambrose doesn't write documentation

#### URL Params

| Name | Type | Description |
| ---- | ---- | ----------- |
| key | String | Audio bite's conversation's ID. |
| start | Epoch timestamp | Time the audio bite starts. |

#### Body

Raw body of audio data in bytes.

#### Success (200 OK)

Empty body.

#### Errors

| Code | Description |
| ---- | ----------- |
| 400 | start is not an uint/key is not an alphanumeric string/data could not be read from the body/bad user claim |
| 500 | Error serialising data into a protocol buffer. |

---

### Put Bite User

```
PUT /conversation/:key/start/:start/user
```

TODO: Description of what this does cos honestly I have no idea Ambrose doesn't write documentation

#### URL Params

| Name | Type | Description |
| ---- | ---- | ----------- |
| key | String | Audio bite's conversation's ID. |
| start | Epoch timestamp | Time the audio bite starts. |

#### Body

Raw body of audio data in bytes.

#### Success (200 OK)

Empty body.

#### Errors

| Code | Description |
| ---- | ----------- |
| 400 | start is not an uint/key is not an alphanumeric string/data could not be read from the body/bad user claim |
| 500 | Error serialising data into a protocol buffer. |
