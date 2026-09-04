# Prompts

AI tooling used: **Claude Code** (Opus), one session, working directly in the
repository.

Below are the prompts I gave, verbatim and in order, grouped by what they were
driving. Typos are left as typed.

## 1. Containerisation

```
create a docker compose to build both web and api
```
```
compose up
```

## 2. Frontend rebuild — macOS calculator

```
Create a simple React web. This should be a calculator so look up for the macos builtin calculator ux design. It fetch the data from the sezzle-calculator-api so add a client for http calls. This web app should have Dockerfile in it. In addition, you can just use prod env. Follow the clean code rules and add unit tests.
```
```
add an input upon the operation input to show the last operation
```
```
1 step- user type the prob as 7x2 and it displays at the below input. And if a
history exists, it shows on the upper input. it history doesnt exist, it should
be empty.
```
```
[screenshot attached] remove the redundant req id input here
```
```
in the web project, we can move the constants KEY_BY_CHARACTER to another file
```
```
follow the owasp 10 and look up if there is a xss
```
```
fix the bugs you caught
```


## 3. Extending the API

```
by following the architecture of the api and web, add this operations as well
Optional: Exponentiation, Square Root, Percentage
```
```
add healthcheck endpoint for liveness
```
```
add gracefully shutdown
```
```
use echo swagger swaggo and add swagger. Refactor the endpoint for the swagger docs. 
```
```
add a Makefile for swagger and mockgen generation. 
```
```
i wrote a service test to show an example. continue to write unit tests. Add non divide zero and overflow tests.
```
```
export the coverage
```
```
add zerelogger and dont use slog. and add a custom http error handler to use
```



## 4. Review and fixes

```
/code-review
```
```
/code-review medium
```
```
fix all
```

## 5. Coverage

```
export the coverage reports
```
```
now run all tests api and web and export new coverage
```
```
exclude docs,configmanager and server and main from coverage
```


## 6. The display overflow bug

```
When a calculation goes on for too long and the numbers get big, they overflow
from the container. Investigate the cause and then we can come up with solutions
ultrathink
```
```
go with A + B and fix the Inf too ultrathink
```
```
can we make the side margins a bit bigger still there is some cutoff when we hit
the limits of the container on the ui
```

## 7. README and deliverables
 
```
[pasted the assignment brief]

tell me your opinions about the project by following these
```
```
Write a README at the root of the folder
```
```
additonaly add docker up instuructions
```
```
keep it simple. just one command for compose is enough
```
