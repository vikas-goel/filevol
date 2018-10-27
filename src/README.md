# docker filevol plugin
Docker Volume Driver for filevol volumes

This plugin can be used to create filevol volumes of specified size, which can
then be bind mounted into the container using `docker run` command.

## Setup

	1) cd docker-filevol-plugin
	2) make
	3) sudo make install

## Usage

1) Start the docker daemon before starting the docker-filevol-plugin daemon.
   You can start docker daemon using command:
```bash
sudo systemctl start docker
```
2) Once docker daemon is up and running, you can start docker-filevol-plugin daemon
   using command:
```bash
sudo systemctl start docker-filevol-plugin
```
NOTE: docker-filevol-plugin daemon is on-demand socket activated. Running `docker volume ls` command
will automatically start the daemon.

3) Add volume path in the config file
```bash
/etc/docker/filevol-plugin
```

## Volume Creation
`docker volume create` command supports the creation of regular volumes, thin volumes, snapshots of regular and thin volumes.

Usage: docker volume create [OPTIONS] [VOLUME]
```bash
-d, --driver    string    Specify volume driver name (default "local")
--label         list      Set metadata for a volume (default [])
-o, --opt       map       Set driver specific options (default map[])
```
Following options can be passed using `-o` or `--opt`
```bash
--opt size
--opt source
```
Please see examples below on how to use these options.

## Examples
```bash
$ docker volume create -d filevol --opt size=208896 foobar
```
This will create a filevol volume named `foobar` of size 208 MB (0.2 GB).
```bash
docker volume create -d filevol --name foobar_snapshot --opt source=foobar
```
This will create a snapshot volume of `foobar` named `foobar_snapshot`.

## Volume List
Use `docker volume ls --help` for more information.

``` bash
$ docker volume ls
```
This will list volumes created by all docker drivers including the default driver (local).

## Volume Inspect
Use `docker volume inspect --help` for more information.

``` bash
$ docker volume inspect foobar
```
This will inspect `foobar` and return a JSON.
```bash
[
    {
        "Driver": "filevol",
        "Labels": {},
        "Mountpoint": "/var/lib/docker-filevol-plugin/foobar",
        "Name": "foobar",
        "Scope": "local"
    }
]
```

## Volume Removal
Use `docker volume rm --help` for more information.
```bash
$ docker volume rm foobar
```
This will remove filevol volume `foobar`.

## Bind Mount filevol volume inside the container

```bash
$ docker run -it -v foobar:/home fedora /bin/bash
```
This will bind mount the logical volume `foobar` into the home directory of the container.
