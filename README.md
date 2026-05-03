# ollie-acme

acme interface for ollie. Provides a session list window and per-session chat windows, all driven through the 9P filesystem.

## Build

```sh
cd acme && mk
```

## Usage

Run `ollie-acme` from within acme. It opens an `ollie/sessions` window listing active sessions. Click a session ID to open its chat window. Tag commands:

| Command | Action |
|---|---|
| New | Open `s/new` for editing |
| Kill *id* | Kill a session |
| Refresh | Reload the session list |

Chat windows stream new output via `statewait` and expose:

| Command | Action |
|---|---|
| Prompt | Open the session's `prompt` file for editing |
| Stop | Interrupt the current turn |
| Ctl | Open the session's `ctl` file |

## Kmpl

`acme/scripts/Kmpl` is an acme script for AI code completion. It reads the text before and after the cursor (dot), sends it to `u/complete`, and inserts the result at the cursor position.

Add `Kmpl` to your acme tag or middle-click it to trigger a completion. It uses the `OLLIE_COMPLETE_BACKEND` and `OLLIE_COMPLETE_MODEL` environment variables (same as `u/complete`).

Override per invocation:

```
Kmpl -model phi-4-mini -backend ollama
```
