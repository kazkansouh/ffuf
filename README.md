```
        /'___\  /'___\           /'___\
       /\ \__/ /\ \__/  __  __  /\ \__/
       \ \ ,__\\ \ ,__\/\ \/\ \ \ \ ,__\
        \ \ \_/ \ \ \_/\ \ \_\ \ \ \ \_/
         \ \_\   \ \_\  \ \____/  \ \_\
          \/_/    \/_/   \/___/    \/_/
```

# ffuf - Fuzz Faster U Fool

A fast web fuzzer written in Go.

## kazkansouh updates

This fork of the excellent `ffuf` tool. It is mainly for personal
use. I have made a number of small changes to the code base to make it
easier for me to use. These include:

* parameter encoding (such as `urlencode`, `base64`, `printf`, `json`
  string escapes).
* fuzzing common extensions for discovered files (e.g. if `index.php`
  is found, it will check for `index.php~` as well) but only if `FUZZ`
  is the last parameter of the url.
* changed the output format to place the file name at end of line -
  this is because long file names can screw up the formatting.
* enabled colour output by default -- why would this not be the case!
* when `ctrl-c` is pressed, print a summary line - far too often I
  have cancelled a long running fuzz'ing processes with no knowledge
  of how far through the word list is got.
* small tweaks to request retry and timeouts that appear to work
  better in the environments that I use the tool.
* hide fuzz banner by default - I dont need to see this for the
  10,000th time, and it gets in the way when copying output into
  notes.
* some sanity checks relating to `Content-Type` header - its too easy
  to forget to specify this header and for the fuzzing to fail.
* refuse to overwrite an output file - sadly I have done this far too
  often.
* when using multiple keywords, print them in a consistent order in
  the results.

## Fork status

At current I have not pulled any recent changes from the main repo. If
there are any bug fixes or cool features that matter to me, Ill pull
them.

I dont actively work on this, and only implement features when I need
them.

Also currently no plans exist to submit any changes to the main repo
as the code bases have started to diverge and I dont have the time
dedicate to this.

## Encoder Usage

A common use case for encoders is dealing with passwords:

```plaintext
$ ffuf -w passwords.txt:FUZZ:urlenc -u http://some.domain/login -d 'user=admin&password=FUZZ'
Encountered error(s): 1 errors occured.                                                     
        * No Content-Type header defined, ignore this error with flag: -dont-check-content-type 
...snip...
```

Here it warns that the content type header is missing and refuses to run. So lets try again:

```plaintext
$ ffuf -w passwords.txt:FUZZ:urlenc -u http://some.domain/login -d 'user=admin&password=FUZZ' -H 'Content-Type: application/x-www-form-urlencoded
'
```

To find out what encoders are currently implemented, use encoder list parameter `-el`:

```plaintext
$ ffuf -el
Available encoders:
  urlenc                   | Url encoding. Set 'urlenc_chars=...' to define encoded chars.
  b64                      | Base64. Set 'b64_url=true' for url variant and 'b64_nopad=true' to remove '=' padding.
  json                     | JSON string. Set 'html_escape=true' to encode html chars '<', '&', '>'
  printf                   | Use printf. E.g. 'printf_fmt=0x%05x' for int inputs, or 'printf_fmt=%05s' and 'printf_int=false'
```

Each encoder supports additional parameters. These are given using the
`-ep` parameter.

As an example, consider the common usage to fuzz formatted
numbers. Also, notice that the fuzzed keyword now appears at the end
of the line instead of the beginning.

```plaintext
$ ffuf -w <(seq 1 3):FUZZ:urlenc:printf -ep FUZZ_printf_fmt=%02d -u 'https://postman-echo.com/get?number=FUZZ'
[Status: 200, Size: ····293, Words: ····5, Lines: ····1] 01
[Status: 200, Size: ····293, Words: ····5, Lines: ····1] 03
[Status: 200, Size: ····293, Words: ····5, Lines: ····1] 02
```

Or even dates:

```plaintext
$ ffuf -w <(seq 1 3):DAY:urlenc:printf -ep DAY_printf_fmt=%02d -w <(seq 1 3):MON:urlenc:printf -ep MON_printf_fmt=%02d -u 'https://postman-echo.com/get?date=2021-MON-DAY'
[Status: 200, Size: 305, Words: 5, Lines: 1]
    * DAY: 01
    * MON: 01

[Status: 200, Size: 305, Words: 5, Lines: 1]
    * DAY: 02
    * MON: 02
...snip...
```

In the above, the keywords `DAY` and `MON` are sorted. So they will
always be printed in the same order.

Finally, its also possible to stack encoders - but I have almost never
needed this:

```plaintext
$ ffuf -w <(seq 1 3):FUZZ:urlenc:printf:b64 -u 'https://postman-echo.com/get?number=FUZZ'
[Status: 200, Size: ····305, Words: ····5, Lines: ····1] MHgwMDAz
[Status: 200, Size: ····305, Words: ····5, Lines: ····1] MHgwMDAx
[Status: 200, Size: ····305, Words: ····5, Lines: ····1] MHgwMDAy
```

# Below is original Fuff readme untouched

## Installation

- [Download](https://github.com/ffuf/ffuf/releases/latest) a prebuilt binary from [releases page](https://github.com/ffuf/ffuf/releases/latest), unpack and run!
  or
- If you have go compiler installed: `go get github.com/ffuf/ffuf`

The only dependency of ffuf is Go 1.11. No dependencies outside of Go standard library are needed.

## Example usage

### Typical directory discovery

[![asciicast](https://asciinema.org/a/211350.png)](https://asciinema.org/a/211350)

By using the FUZZ keyword at the end of URL (`-u`):

```
ffuf -w /path/to/wordlist -u https://target/FUZZ
```

### Virtual host discovery (without DNS records)

[![asciicast](https://asciinema.org/a/211360.png)](https://asciinema.org/a/211360)

Assuming that the default virtualhost response size is 4242 bytes, we can filter out all the responses of that size (`-fs 4242`)while fuzzing the Host - header:

```
ffuf -w /path/to/vhost/wordlist -u https://target -H "Host: FUZZ" -fs 4242
```

### GET parameter fuzzing

GET parameter name fuzzing is very similar to directory discovery, and works by defining the `FUZZ` keyword as a part of the URL. This also assumes an response size of 4242 bytes for invalid GET parameter name.

```
ffuf -w /path/to/paramnames.txt -u https://target/script.php?FUZZ=test_value -fs 4242
```

If the parameter name is known, the values can be fuzzed the same way. This example assumes a wrong parameter value returning HTTP response code 401.

```
ffuf -w /path/to/values.txt -u https://target/script.php?valid_name=FUZZ -fc 401
```

### POST data fuzzing

This is a very straightforward operation, again by using the `FUZZ` keyword. This example is fuzzing only part of the POST request. We're again filtering out the 401 responses.

```
ffuf -w /path/to/postdata.txt -X POST -d "username=admin\&password=FUZZ" -u https://target/login.php -fc 401
```

### Maximum execution time

If you don't want ffuf to run indefinitely, you can use the `-maxtime`. This stops __the entire__ process after a given time (in seconds).

```
ffuf -w /path/to/wordlist -u https://target/FUZZ -maxtime 60
```

When working with recursion, you can control the maxtime __per job__ using `-maxtime-job`. This will stop the current job after a given time (in seconds) and continue with the next one. New jobs are created when the recursion functionality detects a subdirectory.

```
ffuf -w /path/to/wordlist -u https://target/FUZZ -maxtime-job 60 -recursion -recursion-depth 2
```

It is also possible to combine both flags limiting the per job maximum execution time as well as the overall execution time. If you do not use recursion then both flags behave equally.

### Using external mutator to produce test cases

For this example, we'll fuzz JSON data that's sent over POST. [Radamsa](https://gitlab.com/akihe/radamsa) is used as the mutator.

When `--input-cmd` is used, ffuf will display matches as their position. This same position value will be available for the callee as an environment variable `$FFUF_NUM`. We'll use this position value as the seed for the mutator. Files example1.txt and example2.txt contain valid JSON payloads. We are matching all the responses, but filtering out response code `400 - Bad request`:

```
ffuf --input-cmd 'radamsa --seed $FFUF_NUM example1.txt example2.txt' -H "Content-Type: application/json" -X POST -u https://ffuf.io.fi/ -mc all -fc 400
```

It of course isn't very efficient to call the mutator for each payload, so we can also pre-generate the payloads, still using [Radamsa](https://gitlab.com/akihe/radamsa) as an example:

```
# Generate 1000 example payloads
radamsa -n 1000 -o %n.txt example1.txt example2.txt

# This results into files 1.txt ... 1000.txt
# Now we can just read the payload data in a loop from file for ffuf

ffuf --input-cmd 'cat $FFUF_NUM.txt' -H "Content-Type: application/json" -X POST -u https://ffuf.io.fi/ -mc all -fc 400
```

## Usage

To define the test case for ffuf, use the keyword `FUZZ` anywhere in the URL (`-u`), headers (`-H`), or POST data (`-d`).

```
Fuzz Faster U Fool - v1.0

HTTP OPTIONS:
  -H               Header `"Name: Value"`, separated by colon. Multiple -H flags are accepted.
  -X               HTTP method to use (default: GET)
  -b               Cookie data `"NAME1=VALUE1; NAME2=VALUE2"` for copy as curl functionality.
  -d               POST data
  -r               Follow redirects (default: false)
  -recursion       Scan recursively. Only FUZZ keyword is supported, and URL (-u) has to end in it. (default: false)
  -recursion-depth Maximum recursion depth. (default: 0)
  -replay-proxy    Replay matched requests using this proxy.
  -timeout         HTTP request timeout in seconds. (default: 10)
  -u               Target URL
  -x               HTTP Proxy URL

GENERAL OPTIONS:
  -V               Show version information. (default: false)
  -ac              Automatically calibrate filtering options (default: false)
  -acc             Custom auto-calibration string. Can be used multiple times. Implies -ac
  -c               Colorize output. (default: false)
  -maxtime         Maximum running time in seconds for the entire process. (default: 0)
  -maxtime-job     Maximum running time in seconds per job. (default: 0)
  -p               Seconds of `delay` between requests, or a range of random delay. For example "0.1" or "0.1-2.0"
  -s               Do not print additional information (silent mode) (default: false)
  -sa              Stop on all error cases. Implies -sf and -se. (default: false)
  -se              Stop on spurious errors (default: false)
  -sf              Stop when > 95% of responses return 403 Forbidden (default: false)
  -t               Number of concurrent threads. (default: 40)
  -v               Verbose output, printing full URL and redirect location (if any) with the results. (default: false)

MATCHER OPTIONS:
  -mc              Match HTTP status codes, or "all" for everything. (default: 200,204,301,302,307,401,403)
  -ml              Match amount of lines in response
  -mr              Match regexp
  -ms              Match HTTP response size
  -mw              Match amount of words in response

FILTER OPTIONS:
  -fc              Filter HTTP status codes from response. Comma separated list of codes and ranges
  -fl              Filter by amount of lines in response. Comma separated list of line counts and ranges
  -fr              Filter regexp
  -fs              Filter HTTP response size. Comma separated list of sizes and ranges
  -fw              Filter by amount of words in response. Comma separated list of word counts and ranges

INPUT OPTIONS:
  -D               DirSearch wordlist compatibility mode. Used in conjunction with -e flag. (default: false)
  -e               Comma separated list of extensions. Extends FUZZ keyword.
  -ic              Ignore wordlist comments (default: false)
  -input-cmd       Command producing the input. --input-num is required when using this input method. Overrides -w.
  -input-num       Number of inputs to test. Used in conjunction with --input-cmd. (default: 100)
  -mode            Multi-wordlist operation mode. Available modes: clusterbomb, pitchfork (default: clusterbomb)
  -request         File containing the raw http request
  -request-proto   Protocol to use along with raw request (default: https)
  -w               Wordlist file path and (optional) keyword separated by colon. eg. '/path/to/wordlist:KEYWORD'

OUTPUT OPTIONS:
  -debug-log       Write all of the internal logging to the specified file.
  -o               Write output to file
  -od              Directory path to store matched results to.
  -of              Output file format. Available formats: json, ejson, html, md, csv, ecsv (default: json)

EXAMPLE USAGE:
  Fuzz file paths from wordlist.txt, match all responses but filter out those with content-size 42.
  Colored, verbose output.
    ffuf -w wordlist.txt -u https://example.org/FUZZ -mc all -fs 42 -c -v

  Fuzz Host-header, match HTTP 200 responses.
    ffuf -w hosts.txt -u https://example.org/ -H "Host: FUZZ" -mc 200

  Fuzz POST JSON data. Match all responses not containing text "error".
    ffuf -w entries.txt -u https://example.org/ -X POST -H "Content-Type: application/json" \
      -d '{"name": "FUZZ", "anotherkey": "anothervalue"}' -fr "error"

  Fuzz multiple locations. Match only responses reflecting the value of "VAL" keyword. Colored.
    ffuf -w params.txt:PARAM -w values.txt:VAL -u https://example.org/?PARAM=VAL -mr "VAL" -c

  More information and examples: https://github.com/ffuf/ffuf
```

## Helper scripts and advanced payloads

See [ffuf-scripts](https://github.com/ffuf/ffuf-scripts) repository for helper scripts and payload generators
for different workflows and usage scenarios.

## License

ffuf is released under MIT license. See [LICENSE](https://github.com/ffuf/ffuf/blob/master/LICENSE).
