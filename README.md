# gsproxy

gsproxy is a proxy between Google Cloud Storage and local.

Inspired by [groovenauts/magellan-gcs-proxy](https://github.com/groovenauts/magellan-gcs-proxy).

## Usage

```
$ gsproxy -key-file=KEY_FILE COMMAND gs://SRC_BUCKET/OBJECT gs://DEST_BUCKET
```

```
$ gsproxy -key-file=key-file.json /bin/cp gs://path/to/bucket/test.txt gs://path/to/bucket/out/test.txt
```

The above will copy `gs://path/to/bucket/in/test.txt` to `gs://path/to/bucket/out/test.txt` .

### Inside

- Make 2 directories in the temporary directory:
    - Directory named `path/to/bucket/in` (source directory).
    - Directory named `path/to/bucket/out` (destination directory).
- Download `gs://path/to/bucket/test.txt` to the source directory.
- Execute COMMAND with arguments:
    1. Downloaded file path.
    2. Destination directory path.
- Upload files in the destination directory to `gs://path/to/bucket/out/test.txt`.