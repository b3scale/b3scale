# Backend Lifecycle Management

## Adding a backend

A backend is either autoregistered (by passing `-a` to `b3scaleagent`), ending up in `init` state.
If autoregistration is disabled, a node can be added manually (see "Adding a backend").

```bash
b3scalectl --api https://api.bbb.example.org add backend https://node23.bbb.example.org/bigbluebutton/api/
```

### Assigning tags

You can assign tags to a node to ensure that only certain frontents can access certain backends. This
way, you can assign pilot customers to nodes running newer BBB versions or grant certain frontends access to nodes with more resources:

```bash
b3scalectl --api https://api.bbb.example.org set backend -j '{"tags":["bbb_26"]}' https://node23.bbb.example.org
```
## Listing backends

You can get a list of all backends including health parameters:

```bash
b3scalectl --api https://api.bbb.example.org show backends

65959a9d-0d1d-476b-ad7a-f665feb63d01
  Host:	 https://node22.bbb.example.org/bigbluebutton/api/
  Settings:	 {"tags":["bbb_25"]}
  NodeState:     ready	  AdminState:	 ready
  MC/AC/R:	 0/0/0.00
  LoadFactor:	 1
  Latency:	 92.5466361ms


6a2f1953-6db3-4efd-bd35-b7d74c1ddd68
  Host:	 https://node23.bbb.example.org/bigbluebutton/api/
  Settings:	 {"tags":["bbb_26"]}
  NodeState:	 ready	  AdminState:	 ready
  MC/AC/R:	 0/0/0.00
  LoadFactor:	 1
  Latency:	 93.076423ms

8bf64dca-18bf-4bb4-9159-2a3e52e7e05b
  Host:	 https://node24.bbb.example.org/bigbluebutton/api/
  Settings:	 {"tags":["bbb_27"]}
  NodeState:	 ready	  AdminState:	 ready
  MC/AC/R:	 0/0/0.00
  LoadFactor:	 1
  Latency:	 91.648361ms
```

## Setting tags

```bash
b3scalectl --api https://api.bbb.example.org set frontend -j ' {"tags":["bbb_28"]}' https://node22.bbb.example.org/bigbluebutton/api/
```

### Abbreviations

* `MC`: Meeting Count
* `AC`: Attendee Count
* `R`: Ratio

## Backend states

Backend nodes in b3scale can be in either of the following state:

* **init**: The node was freshly initialized
* **ready**: The node and is ready for use
* **stopped**: The node has been disabled and is stopped
* **error**: An error has occured
* **decommissioned**: The node has been decomissioned

The current state is expressed in the `NodeState`, which is the acutal state a
node. The `AdminState` however is the *desired*  state, usually mandated by an
administrative change. The two might not be identical all the time, e.g. when a node
loses connection to b3scale, the `NodeState` will be `error`, but the AdminState
will still be `ready`. Since `error` is not a desirable state, AdminState cannot be
`error`.

## Enabling a backend

```bash
b3scalectl --api https://api.bbb.example.org enable backend https://node23.bbb.example.org
```

## Cordoning a backend

```bash
b3scalectl --api https://api.bbb.example.org disable backend https://node23.bbb.example.org
```
!!! note
    Disabling a backend will keep existing meeting in place and active until the meeting has been finished.

## Reintegrating a backend

See section "Enabling a backend".

## Removing a backend

```bash
b3scalectl --api https://api.bbb.example.org delete backend https://node23.bbb.example.org
```

If the node has been removed prior to deregistering it with b3scale, you may need to specify
the `--force` parameter to forcefully remove the backend. It is recommended to always decomission backends through a cordoning phase in order to not disrupt running meetings.
