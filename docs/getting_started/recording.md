# Setting up recording

b3scale can publish recordings recoded on a node and will allow frontends to manage them in three ways:

* Publish a recording
* Unpublish a recording
* Delete a recording
## How it works

### On the b3scale server

b3scale expects two directories with full write access to the `b3scaled` process, passed to the process as environment variables:

* `B3SCALE_RECORDINGS_PUBLISHED_PATH` points to a directory where all published talks will be hosted.
* `B3SCALE_RECORDINGS_UNPUBLISHED_PATH` points to a directory where all unpublished/depublished talks will be hosted

The playback URL will be created using the host provided via the `B3SCALE_RECORDINGS_PLAYBACK_HOST` variable. The playback HTTPS host needs to be set up separately and requires (read-only) access to `B3SCALE_RECORDINGS_PUBLISHED_PATH`.

Recordings can also be configured using the `B3SCALE_RECORDINGS_INBOX_PATH`
to use a dedicated recordings inbox, where
new recordings will be initially placed before importing.

If the `B3SCALE_RECORDINGS_INBOX_PATH` is not set, either
`B3SCALE_RECORDINGS_PUBLISHED_PATH` or `B3SCALE_RECORDINGS_UNPUBLISHED_PATH`
will be used, depending on `B3SCALE_RECORDINGS_DEFAULT_VISIBILITY`.

!!! note
    `B3SCALE_RECORDINGS_DEFAULT_VISIBILITY` only informs b3scale
    where to expect the recording before import and does not affect the actual
    visibility.


### Managing recordings

Using `b3scalectl` recordings associated with a frontend can be listed
with

```bash
b3scalectl show recordings --frontend example-frontend-key
```
you can also enable `json` output:

```bash
b3scalectl show --json recordings --frontend example-frontend-key
```

By listing the recordings you can get the ID of a recording.
With a known recording ID, you can change the visibility:

```bash
b3scalectl set recording-visiblity <id> <visiblity>
```

where `<visibility>` may be anything of `published`, `protected`, `public`,
`public_protected` and `unpublished`.

### Setting default visibility policy for a Frontend

If a frontend is incapable of handling recording visibilities, b3scale allows to override
the default visibility per frontend by specifying the `recordings.visibility_override` property:

```bash
b3scalectl set frontend -j '{"recordings": { "visibility_override": "protected" }}' <frontend>
```

!!! note
    This will not change the visibility of existing recording. See the section above on
    how to modify existing recordings.

### On the BigBlueButton node

Publishing the recordings requires a post-publish hook script to be placed in the in `/usr/local/bigbluebutton/core/scripts/post_publish`folder on all BBB nodes.

Depending on your setup you either need to copy (via rsync or scp) to `B3SCALE_RECORDINGS_PUBLISHED_PATH` on a node that has access to that location, or mount a shared volume.

Finally, the `metadata.xml` document, created by the BigBlueButton node for the respective recording, needs to be submitted to `recording-import` API endpoint on b3scale.

See `examples/post_publish_b3scale_import.rb` for a trivial example that assumes that all nodes, the b3scale server and the playback host share the same `/var/bigbluebutton` directory via a shared filesystem such as NFS or CephFS.
