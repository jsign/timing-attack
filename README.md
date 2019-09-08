# Timing-attack

[Timing-attack](https://en.wikipedia.org/wiki/Timing_attack) is a proof of concept of a type of side-channel attack where time metadata is use to leak information about executions. In particular, network API calls are explored to identify tiny latencies variations to assert with high-confidence some facts about different request cases.

# Design
This proof of concept is separated in two parts: _server_ and _attacker_.

## Server
The _server_ is a simple webserver mock which simulates an API login, where first the user is retrieved by email from a database, and then, if exists, a password checksum is done. The database query and the checksum comparation are mocked by different `time.Sleep` calls with reasonable order-magnitude values.

Also, base-latencies and standar-deviation values can be configured to add noise to the total request latency to make it more realistic.

## Attacker
The _attacker_ is a tool which using multiple test cases of emails for logins, makes an statistical analysis to conclude if a particular test-case has a meaningful statistical difference in latency to conclude something about the implementation runtime path.

# Usage
Terminal 1:
```
timing-attack [master] % make run-server
```
Terminal 2:
```
timing-attack [master] % make run-attacker
DEBU[0004] Max median latency: correct@email.com in 17.17ms 
DEBU[0004] Base average latency is: 16.06ms             
DEBU[0004] Base stddev is: 5.15ms                       
DEBU[0004] Median latency for correct@email.com is 17.17ms (21.54%) 
DEBU[0004] Median latency for whatever@fake.com is 15.94ms (-2.44%) 
DEBU[0004] Median latency for foo@fake.com is 15.94ms (-2.39%
```

# Further work
This is just an initial exploration of the problem. Further work might improve significantly precision, performance, configurability, and enabling to use attacker as a library rather than a tool.

# License
timing-attack is licenced under the MIT license.
