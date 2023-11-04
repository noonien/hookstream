# hookstream

hookstream listens for webhooks and broadcasts the data to websocket listeners. It's designed to handle webhooks on a per-topic basis, offering both dedicated topic streams as well as a broadcast channel for all topics.

## Features

- **Per-Topic Webhooks**: Receive webhooks at `/hook/{topic}` and broadcast to websocket listeners on `/socket/{topic}`.
- **Global Broadcast**: Send received hooks to all connected websocket clients on `/socket` regardless of the topic.
- **Simple Integration**: Easy to integrate into your existing system with minimal configuration.
- **WebSocket Support**: Leveraging the robustness of websockets for real-time data streaming.
- **JSON detection**: If the webhook sends a JSON, it's directly embedded in the broadcast message, instead of encoding it as a string.
## Install

```sh
go install github.com/noonien/hookstream@latest
```

Prebuilt binaries are available on the [Release page](https://github.com/noonien/hookstream/releases).

## Usage

```sh
hookstream -h
Usage of hookstream:
  -addr string
    	address to listen on (default ":8080")
  -prefix string
    	http prefix to serve on (default "/")
```

Once hookstream is running, you can set up your HTTP hooks to send requests to:

```
http://<your-server-address>/hook/{topic}
```

Where `{topic}` is the identifier for your specific stream.

Listeners can connect to the websocket endpoint at:

```
ws://<your-server-address>/socket/{topic}
```

For global broadcasts (all topics), websocket clients can connect to:

```
ws://<your-server-address>/socket
```

When running behind a reverse proxy, use the `-prefix` argument to serve at a non-root path.

## Example

Run the server:
```sh
hookstream -addr :8080
```

Connect to `ws://localhost:8080/socket/mytopic` using a websocket client:
```sh
websocat ws://localhost:8080/foo/socket/mytopic
```

Use cURL to send data to the hookstream server:

```sh
curl -X POST http://localhost:8080/hook/mytopic -d '{"message": "Hello, World!"}'
```

The websocket client will receive the following data, formatted for readability:

```json
{
  "topic": "mytopic",
  "method": "POST",
  "headers": {
    "Accept": "*/*",
    "Content-Length": "28",
    "Content-Type": "application/x-www-form-urlencoded",
    "User-Agent": "curl/7.86.0"
  },
  "data": {
    "message": "Hello, World!"
  }
}
```

### Authentication

Authentication is not handled natively. It is recommended to use a reverse proxy to manage authentication and secure your hookstream server.

## License

hookstream is released under the MIT License. See the `LICENSE` file for more details.
