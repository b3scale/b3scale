# Frontend Lifecycle Management


## Adding a frontend

```bash
b3scalectl --api https://api.bbb.example.org add frontend --secret "$(pwgen -s 1 42)" my-frontend
```

## Assigning tags and properties to a frontend

```bash
b3scalectl --api https://api.bbb.example.org set frontend -j '{"required_tags":["bbb_26"]}' my-frontend
```

You can also specify the `-j` parameter while adding a frontend with `add frontend`.

### Valid properties

#### `required_tags`

Associates one or more frontends with one or more backends with the same tags. Matches the `{"required_tags":["bbb_26"]}` specification in backend properties.
#### `default_presentation`

```JSON
"default_presentation":{"url":"https://static.example.com/presentation/my-frontend.pdf","force":false}
```

Specifies a default presentation for the tenant, which will be downloaded from `url`. Specify `"force":true` to ignore any custom presentation uploaded via the frontend application.

###  `create_default_params`

Parameters ([see documentation](https://docs.bigbluebutton.org/development/api#create)) to be set if the frontend application does not specify the parameter. For example, the following will set the default welcome message to "Welcome to you meeting!".

```JSON
"create_default_params":{"welcome":"Welcome to your meeting!"}
```
###  `create_override_params`

Parameters ([see documentation](https://docs.bigbluebutton.org/development/api#create)) that will override the respective parameter of the frontend application. For example, the following will ensure that the learning dashboard is disabled and recording is off:

```JSON
"create_override_params":{"disabledFeatures":"learningDashboard","record":"false"}
```

## Using a frontend

To test the frontend, you can use `https://mconf.github.io/api-mate/`. Use `https://api.bbb.example.org/bbb/my-frontend/bigbluebutton/api` as the link and the secret. You can also use this URL for Greenlight or other frontends.

## Listing frontends

```bash
b3scalectl --api https://api.bbb.example.org show frontends

37dada5b-9fd0-41b5-8875-17037a4d5413 my-frontend eefeng2aixaeghu0tae9ja4ietheeNgoo5ubunga1ohkeexaib3bai9xudaang2i {"required_tags":["bbb_26"]}
...
```
## Modifying a frontend

You can use `set frontend` with `-j` (see "Assigning tags to a frontend") or `--secret` (see "Adding a frontend").

## Removing a frontend

```bash
b3scalectl --api https://api.bbb.example.org delete frontend my-frontend
```