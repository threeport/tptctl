# Commands & Configs

We are following the command structure as advocated by in the [Cobra
docs](https://cobra.dev/#concepts).  This will make the CLI tool as intuitive as
possible for users.

The pattern is:

```bash
tptctl verb noun --adjective
```
or

```bash
tptctl command arg --flag
```

## Commands

Using this command pattern allows us to reduce the number of commands by
applying the same verb notion to many different things.

### Create Command

Following are different example usages of the `create` command.

The create command can create three different types of objects:

* Control Plane Instances: a Threeport control plane
* Objects: a single object defined by a distinct `struct` and with an API
  endpoint, e.g. `WorkloadDefinition`
* Constructs: abstractions that included multiple objects and have an API
  endpoint, e.g. `Workload`

Create a Threeport control plane.  When you create an instance of Threeport,
tptctl will write (or update) a file to your local filesystem that contains
the superuser credentials to that instance, by default at
`~/.config/threeport/config.yaml`.

```bash
tptctl create control-plane \
    --provider kind \  # required
    --name dev \  # required
    --threeport-config-file-out /non/default/location/config.yaml  # optional
```

Create a single API object in an instance of Threeport:

```bash
tptctl create workload-definition \
    --config-file /tmp/workload-def.yaml \  # required
    --instance dev \  # optional (defaultable in threeport config file)
    --credentials rbac-test \  # optional (defaultable in threeport config file)
    --threeport-config-file /non/default/location/config.yaml  # optional (default: ~/.config/threeport/config.yaml)
```

Create an abstracted construct of multiple objects.  The following creates a
workload definition and a workload instance with one command:

```bash
tptctl create workload \
    --config-file /tmp/workload.yaml
```

Create an object with the object declared in the config file:

```bash
tptctl create object \
    --config-file /tmp/object.yaml
```

### Delete Command

The delete command is simply the converse of create.

Delete an instance of Threeport:

```bash
tptctl delete threeport \
    --name dev
```

Delete a workload definition object by name:

```bash
tptctl delete workload-definition \
    --name "web3-sample-app"
```

Deleting constructs is a little treacherous since there are multiple associated
objects.  For this reason, constructs cannot be deleted through tptctl - or
the Threeport API for that matter.

## Config Files

There are two general classes of config file:

* Threeport config: used by tptctl to connect to Threeport API.
* Object/Construct config: used to define attibutes of objects and constructs.

### Threeport Config

The following includes configuration for to different Threeport instances, one
called "prod," the other called "dev."  The "dev" instance includes credentials
for two different users.  Anything configuration that defines an array of
objects has a `Default` field.

```yaml
Instances:
  - Name: "prod"
    Tags:
      - tier: "prod"
      - owner: "bob"
    APIServer: "https://os.qleet.io"
    Credentials:
      - Name: "superuser"
        Tags:
          - tier: "prod"
        Email: "richard@qleet.io"
        Password: "c2VjcmV0Cg=="
        Default: true
    Default: false
  - Name: "dev"
    APIServer: "http://localhost:1323"
    Credentials:
      - Name: "superuser"
        Email: "richard@qleet.io"
        Password: "Zm9vCg=="
        Default: true
      - Name: "rbac-test"
        Email: "platform-engineering@qleet.io"
        Password: "Zm9vCg=="
        Default: "false"
    Default: true
```

### Object Config

The following is a configuration for a workload definition:

```yaml
Name: "web3-sample-app"
YAMLDocument: "/tmp/resources.yaml"
UserID: 1
```

#### Consideration & Proposal

We don't allow the creation of multiple objects when calling object endpoints
Should we allow that through tptctl?

Proposal:

* We retain this existing policy in the API
    - e.g. calling `{{host}}:{{port}}/v0/users` allows only one user at a time
* We _do_ allow creation of multiple objects through "construct endpoints"
    - e.g. can create multiple users at `{{host}}:{{port}}/v0/usersets`
* We provide the opportunity to make create multiple objects with a single
  command through tptctl by either a) making multiple API calls or b) calling
  construct endpoints or c) both

The following is a configuration for a workload definition that references a
user by email instead of ID.

```yaml
Name: "web3-sample-app"
YAMLDocument: "/tmp/resources.yaml"
User:
  Email: "richard@qleet.io"  # must be field tagged gorm:"unique"
```

The following is a configuration for a workload definition that creates a user
and a workload definition with one command.  This is an example where tptctl
will make two API calls.  The user must define dependenet objects first in the
config file:

```yaml
Email: "bob@org.com"
Password": "asdf"
FirstName": "Bob"
LastName": "Smith"
DateOfBirth": "1985-01-30T00:00:00Z",
CountryOfResidence": "United States",
Nationality": "United States"
---
Name: "web3-sample-app"
YAMLDocument: "/tmp/resources.yaml"
User:
  Email: "bob@org.com"
```

An optional Object field can allow the command to omit the object name:

```yaml
Object: WorkloadDefinition
Name: "web3-sample-app"
YAMLDocument: "/tmp/resources.yaml"
User:
  Email: "richard@qleet.io"  # must be field tagged gorm:"unique"
```

A Construct field can allow a construct to be declared in the config file:

```yaml
Construct: Workload
WorkloadDefinition:
  Name: "web3-sample-app"
  YAMLDocument: "/tmp/resources.yaml"
  User:
    Email: "richard@qleet.io"  # must be field tagged gorm:"unique"
WorkloadInstance:
  Name: "web3-sample-app"
  WorkloadCluster:
    Name: "dev-01"
```

An optional Version field can be used to use a previous version of an API:

```yaml
Version: v0
Name: "web3-sample-app"
YAMLDocument: "/tmp/resources.yaml"
User:
  Email: "bob@org.com"
```

Optional Group and Object fields can be used for custom APIs:

```yaml
Group: eth
Object: Node
Name: "eth-node"
Tier: "prod"
```

