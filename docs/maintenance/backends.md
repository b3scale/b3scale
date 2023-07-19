# Backend Lifecycle Management

## Listing backends

```bash
b3scalectl --api https://api.bbb.example.org show frontends

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
  MC/AC/R:	 0/2/0.00
  LoadFactor:	 1
  Latency:	 91.648361ms
```
## Backend states

NodeState:

* **new**:
* **ready**:
* **error**:


AdminState:

## Adding a backend

```bash
b3scalectl --api https://api.bbb.example.org add backend https://node23.bbb.example.org
```

Usually not needed due to autoregistration. Simply enable the backend (see below).

## Enabling a backend

```bash
b3scalectl --api https://api.bbb.example.org enable backend https://node23.bbb.example.org
```

### Assigning tags

```bash
b3scalectl --api https://api.bbb.example.org set backend -j '{"tags":["bbb_26"]}' https://node23.bbb.example.org
```

## Cordoning a backend

```bash
b3scalectl --api https://api.bbb.example.org disable backend https://node23.bbb.example.org
```

## Reintegrating a backend

See section "Enabling a backend".

## Removing a backend

```bash
b3scalectl --api https://api.bbb.example.org delete backend https://node23.bbb.example.org
```

Accepts `--force`.