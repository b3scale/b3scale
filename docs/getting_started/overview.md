# Overview

## About

`b3scale` is a load balancer designed to be used in place of scalelite. Work was
started in 2020 to provide multiple features not possible before:

  * *API driven*: REST API allows integrating b3scale straight to your CRM
    and/or user portal
  * *Observable*: Prometheus endpoint for all essential operational information
  * *Multi tenancy*: b3scale introduces the concept of frontends
    * custom client secret
    * optional custom presentation
  * *Efficient*: 5 figure users with a single instance
  * *High availability and easy scale out*: Just add more b3scale servers and
    use with your existing HA solution
  * *Flexible backend handling*:
    * Map frontends to BBB nodes (backends)
    * Retire backends for updates
    * Powerful tagging system allows for friendly user testing, experiments and
      assignments depending on expected customer load
  * *True load-awareness*: using reports from the `b3scaleagent`, b3scale
    can schedule meetings more efficiently
  * *Easy deployment*: b3scale is written in Go: no dependencies, just deploy a
    single binary

## Concepts

BigBlueButton arrives with a simple concept: One API, one frontend, usually Greenlight, authenticated through an API secret. As a scaler and
load balancer for BigBlueButton, b3scale expands on that by introducing additional concepts:

* **Node:** An individual installation of BigBlueButton
* **Backend:** A BigBlueButton Node made available as a resource to b3scale
* **Frontend:** A frontend is used to distinguish multiple tenants. It is a tuple of a frontend name, an instance and JSON-encoded settings valid for that frontend.
* **Tags:** Used to associate one or more frontends with one ore more backends (n:m relation)
* **Metrics:** b3scale provides [Prometheus](https://prometheus.io/)-compatible metrics about front- and backends
* **JWTs**: [JSON Web-Tokens](https://jwt.io/) are used to authorize services with b3scale 


b3scale services different *frontends*. Those can be standard apps such as
Greenlight, Nextcloud or Moodle, but can also really be anything that implements
the BigBlueButton API, even custom web apps.

A frontend can initiate a new meeting via b3scale, which will assign it to a
*backend* node and will keep track of the assignment. Users joining will thus
be assigned to the correct backend.

Using *tags* you can assign specific roles to one or more backend nodes: you
can assign a customer to a specific set of nodes, effectively forming a
dedicated cluster. It is also possible to  steer friendly users towards nodes
that contain experimental features.

You can take a backend offline by disabling it. This will not affect currently
running meetings. It will only remove the node from consideration for new
meetings. This way, backend nodes can be drained e.g. in preparation for
scheduled maintenance.

For backends nodes, b3scale provides `b3scaleagent`, an process for backend nodes
that monitors certain parameters straight from redis and reports them to
b3scale in an inexpensive, resource conserving fashion.

The following architecture overview shows these concept coming together:

![b3scale architecture](../assets/images/b3scale-architecture.png#only-light){ loading=lazy }
![b3scale architecture](../assets/images/b3scale-architecture_dark.png#dark-light){ loading=lazy }
<figure markdown>
  <figcaption>b3scale architecture overview</figcaption>
</figure>

## Components

b3scale consists of several components:

* **Database:** b3scale requires a Postgres-Database as storage backend
* **b3scaled:** The central scaling service that accepts requests from frontends and distributes them to backends
* **b3scalectl:** The command line tool, a wrapper
* **b3scaleagent:**  An agent process that reports status and health of a backend to the central scaling service using a REST API
