# `lbtest`

`lbtest` contains basic tools for testing the functionality of a TCP load
balancer.

- `upstreams` runs 1 or more TCP echo servers and tracks how many connections
  each receives
- `client` sends arbitrary data to a particular TCP port and verifies that it is
  echoed back to the client