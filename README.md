# gitlab-workhorse

Gitlab-workhorse is a smart reverse proxy for GitLab. It handles
"large" HTTP requests such as file downloads, file uploads, Git
push/pull and Git archive downloads.

## Quick facts (how does Workhorse work)

-   Workhorse can handle some requests without involving Rails at all:
    for example, Javascript files and CSS files are served straight
    from disk.
-   Workhorse can modify responses sent by Rails: for example if you use
    `send_file` in Rails then gitlab-workhorse will open the file on
    disk and send its contents as the response body to the client.
-   Workhorse can take over requests after asking permission from Rails.
    Example: handling `git clone`.
-   Workhorse can modify requests before passing them to Rails. Example:
    when handling a Git LFS upload Workhorse first asks permission from
    Rails, then it stores the request body in a tempfile, then it sends
    a modified request containing the tempfile path to Rails.
-   Workhorse can manage long-lived WebSocket connections for Rails.
    Example: handling the terminal websocket for environments.
-   Workhorse does not connect to Redis or Postgres, only to Rails.
-   We assume that all requests that reach Workhorse pass through an
    upstream proxy such as NGINX or Apache first.
-   Workhorse does not accept HTTPS connections.
-   Workhorse does not clean up idle client connections.
-   We assume that all requests to Rails pass through Workhorse.

For more information see ['A brief history of
gitlab-workhorse'][brief-history-blog].

## Usage

```
  gitlab-workhorse [OPTIONS]

Options:
  -apiLimit uint
        Number of API requests allowed at single time
  -apiQueueDuration duration
        Maximum queueing duration of requests (default 30s)
  -apiQueueLimit uint
        Number of API requests allowed to be queued
  -authBackend string
    	Authentication/authorization backend (default "http://localhost:8080")
  -authSocket string
    	Optional: Unix domain socket to dial authBackend at
  -developmentMode
    	Allow to serve assets from Rails app
  -documentRoot string
    	Path to static files content (default "public")
  -listenAddr string
    	Listen address for HTTP server (default "localhost:8181")
  -listenNetwork string
    	Listen 'network' (tcp, tcp4, tcp6, unix) (default "tcp")
  -listenUmask int
    	Umask for Unix socket
  -pprofListenAddr string
    	pprof listening address, e.g. 'localhost:6060'
  -proxyHeadersTimeout duration
    	How long to wait for response headers when proxying the request (default 5m0s)
  -secretPath string
    	File with secret key to authenticate with authBackend (default "./.gitlab_workhorse_secret")
  -version
    	Print version and exit
```

The 'auth backend' refers to the GitLab Rails application. The name is
a holdover from when gitlab-workhorse only handled Git push/pull over
HTTP.

Gitlab-workhorse can listen on either a TCP or a Unix domain socket. It
can also open a second listening TCP listening socket with the Go
[net/http/pprof profiler server](http://golang.org/pkg/net/http/pprof/).

### Relative URL support

If you are mounting GitLab at a relative URL, e.g.
`example.com/gitlab`, then you should also use this relative URL in
the `authBackend` setting:

```
gitlab-workhorse -authBackend http://localhost:8080/gitlab
```

## Installation

To install gitlab-workhorse you need [Go 1.5 or
newer](https://golang.org/dl) and [GNU
Make](https://www.gnu.org/software/make/).

To install into `/usr/local/bin` run `make install`.

```
make install
```

To install into `/foo/bin` set the PREFIX variable.

```
make install PREFIX=/foo
```

On some operating systems, such as FreeBSD, you may have to use
`gmake` instead of `make`.

## Error tracking

GitLab-Workhorse supports remote error tracking with
[Sentry](https://sentry.io). To enable this feature set the
GITLAB_WORKHORSE_SENTRY_DSN environment variable.

Omnibus (`/etc/gitlab/gitlab.rb`):

```
gitlab_workhorse['env'] = {'GITLAB_WORKHORSE_SENTRY_DSN' => 'https://foobar'}
```

Source installations (`/etc/default/gitlab`):

```
export GITLAB_WORKHORSE_SENTRY_DSN='https://foobar'
```

## Tests

Run the tests with:

```
make clean test
```

### Coverage / what to test

Each feature in gitlab-workhorse should have an integration test that
verifies that the feature 'kicks in' on the right requests and leaves
other requests unaffected. It is better to also have package-level tests
for specific behavior but the high-level integration tests should have
the first priority during development.

It is OK if a feature is only covered by integration tests.

## License

This code is distributed under the MIT license, see the LICENSE file.

[brief-history-blog]: https://about.gitlab.com/2016/04/12/a-brief-history-of-gitlab-workhorse/
