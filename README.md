# 📚 CO663 Assignment 1: Git Bisect

@ University of Kent

CO663 Assessment 1 - Git Bisect

Here is the project requirements taken from [the homework outline](https://rgrig.github.io/plad/homework.html)

## Outline

Most software projects reside in a Git repository. One of the more useful Git commands is bisect. The following is an example from its manual page:

```bash
git bisect start
git bisect bad                 # Current version is bad
git bisect good abc1234        # abc1234 is known to be good
```

The current version has some bug (is bad); version abc1234 does not (is good). The goal is to find the commit that introduced the bug. Git will checkout some version between abc1234 and the current one and then let you figure out if the bug is present. If the bug is present, you should run the following command:

```bash
git bisect bad
```

If the bug is not present, you should run the following command:

```bash
git bisect good
```

After this, Git checks out another version, you inform Git if the version is bad or good, and this process repeats several times. Eventually, Git reports what is the commit that introduced the bug. Now you have a good hint about where in the code you should look. Or, at least, you know who to blame.

Git tries to figure out where the bug was introduced by asking few questions. After all, figuring out if the bug is present or not may be time consuming: at the very least, it involves compiling the whole project. Git-bisect uses a heuristic that works on arbitrary histories, and degenerates to binary search when used on linear histories. Your task is to try to do better.

### A Small Example

You will write a program that communicates with a server: your program plays the role of git-bisect, the server plays the role of a human. The communication works over WebSockets, with messages in JSON format. We use WebSockets because, unlike sockets, they have a notion of message; that is, communication is not a stream of bytes, but a stream of messages. We use JSON because it is a widely used standard for representing structured information in a human-readable format, which makes it easy to debug.

Let us consider an example. Immediately after connecting, the client authenticates by sending the following message:

```json
{"User": ["rg399", "TOKEN"] }
```

Above, TOKEN should be replaced by the access token you received by email.

Next, the server replies with a repository description (Repo), followed by a problem instance (Instance):

```json
{"Repo":{"name":"pb0","instance_count":10,"dag":[["a",[]],["b",["a"]],["c",["b"]]]}}
{"Instance":{"good":"a","bad":"c"}}
```

The dag (directed acyclic graph) lists the parents of each commit:

a has no parent
b has parent a
c has parent b
This is a linear history, because each commit but one has exactly one parent (i.e., there is no branching). In general, one can have merge commits, which have multiple parents. Cycles, however, will never be present.

The Repo message is followed by an Instance message, which tells us that a is good and c is bad. So, the bug could have been introduced by b or c; but, which one?

The client now asks, and the server responds:

```json
{"Question":"b"}
{"Answer":"Good"}
```

And again:

```json
{"Question":"c"}
{"Answer":"Bad"}
```

At this point, the client figured out that the bug was introduced in version c, and communicates its solution to the server:

```json
{"Solution":"c"}
```

The server now continues in one of three ways:

With a new Repo message.
With a new Instance message. (Each Instance message refers to the last Repo message. The Repo message gives the dag; the Instance message gives the known good/bad commits.)
With a Scores message, if the submission was successful.
In this example, there is no further problem instance to solve, so the server replies with:

```json
{"Score":{"pb0":{"Correct":2}}}
```

This message says that the solution was correct, and it was found with 2 questions. The other possible outcomes are the following:

```json
{"Score":{"pb0":"Wrong"}}
{"Score":{"pb0":"GaveUp"}}
```

The last outcome is achieved by sending a "GiveUp" message instead of a question or a solution.

Observe that, from the problem instance, we already know that c is bad, so the second question asked by the client is redundant: this problem instance can be solved with just one question!

## RUN LOCALLY

The actual submission was made using websockets, but since that server is probably down by the time you read this, you can see the output in [results.txt](results.txt)

Simply run:

```bash
go run cmd/fromwebsockets/main.go
```

And this should run for a while and eventually output your results.

FYI: this was optimised for multiprocessing, so the more CPU's you chuck at this thing, the better it gets. However still remains to be seen if the multiprocessing overhead actually slows it down?
