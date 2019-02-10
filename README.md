# backend-publish

Beep backend accepts PUT requests and publishes a protobuf-ed version to a [NATS](htts://nats.io) queue, like some sort of weird HTTP/NATS converter. Also handles authentication of said HTTP requests. Needless to say, relies on a NATS instance being up.

## Quickstart

```
go build && ./backend-publish
```

## Flags

Flags are supplied to the compiled go program in the form ```-flag=stuff```.

| Flag | Description | Default |
| ---- | ----------- | ------- |
| listen | Port number to listen on | 8080 |
| nats | URL of NATS | nats://localhost:4222 |

## API

All requests require an ```Authorization: Bearer <token>``` header, with token being obtained from ```backend-login```.

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
| 400 | start is not an uint/key is not an alphanumeric string/data could not be read from the body |
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
| 400 | start is not an uint/key is not an alphanumeric string/data could not be read from the body |
| 500 | Error serialising data into a protocol buffer. |
