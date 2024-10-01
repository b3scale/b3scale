# B3scale
*The efficient multi tenant load balancer for BigBlueButton*

[![Test](https://github.com/b3scale/b3scale/actions/workflows/main.yml/badge.svg)](https://github.com/b3scale/b3scale/actions/workflows/main.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/b3scale/b3scale)](https://goreportcard.com/report/github.com/b3scale/b3scale)

## Mission Statement

Efficiently provide access to a (single) pool of BigBlueButton servers to multiple BBB
frontends such as Greenlight or Moodle while at least maintaining feature parity with
[Scalelite](https://github.com/blindsidenetworks/scalelite).

## Feature Matrix

|                                | Scalelite | b3scale |
| ------------------------------ | --------- | ------- |
| Multiple Backends              |     ✅    |    ✅   |
| Multiple Frontends             |     ❌    |    ✅   |
| Customizable Frontend Settings |     ❌    |    ✅ <sup>1)</sup>   |
| Recording Support              |     ✅    |    ✅   |
| Protected Recordings           |     ✅    |    ✅   |
| Predictable Dialin Numbers     |     ✅ <sup>2)</sup> |    🚧 <sup>3)</sup> |
| Frontend agnostic              |     ✅    |    ✅   |
| Agent-based Node Monitoring    |     ❌    |    ✅   |
| Prometheus Exporter            |     ❌    |    ✅   |
| Administration via API         |     ❌    |    ✅   |
| Administration via Web-UI      |     ❌    |    ❌   |
| Administration via CLI         |     ✅ <sup>4)</sup>  |    ✅   |
| Kubernetes-Operator            |     ❌    |    ✅ <sup>5)</sup>   |

<sup>1)</sup> Through overridable/default `create` API parameters or tagged, custom backend servers<br>
<sup>2)</sup> Random, static assignment only<br>
<sup>3)</sup> See https://github.com/b3scale/b3scale/issues/155<br>
<sup>4)</sup> Limited set of commands available via Rake tasks<br>
<sup>5)</sup> Frontend provisioning only

## Documentation

Find user and API documentation, Getting Started guide and more
on the [official b3scale website](https://b3scale.io).

## Bug reports and Contributions

If you discover a problem with b3scale or have a feature request, please open a
[bug report](https://github.com/b3scale/b3scale/issues/new). Please
check the [existing
issues](https://github.com/b3scale/b3scale/issues) before reporting
new ones. Do not start work on new features without prior discussion. This
helps us to coordinate development efforts. Once your feature is discussed,
please file a merge request for the `develop` branch. Merge requests to
`main`happen from `develop` only.

## Discussions

Please use [GitHub Discussions](https://github.com/b3scale/b3scale/discussions)
for Q&A, Feedback, presenting clever solutions and more.

## License

b3scale is provided under the [GNU Affero General Public License 3.0](https://www.gnu.org/licenses/agpl-3.0.en.html).
That means that all changes made to b3scale by an operating party
must be provided as described by the license. Unlike other projects,
contributing to b3scale does not require signing a Contributor Agreement
or similar. This means fair, impartial treatment for the entire community.

## Disclaimer

*This project uses BigBlueButton and is not endorsed or certified by
BigBlueButton Inc. BigBlueButton and the BigBlueButton Logo are trademarks of
BigBlueButton Inc.*
