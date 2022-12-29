# Replaystream

This is primarily a testing tool, but it could likely be used for other purposes like ghost "demo" data.

## Usage

There are four modes of Replaystream.

### Copy

```bash
./replaystream copy -x 83 -a 18 -s 1614951600000 -e 1614951600001 -d "daycare trip in f150 with AVL" -n "truckster 5 min trip" --source ./dev-latinum-16efc73f580c.json -g ./test-latinum-3cba82351b2d.json -t avl_status -t e2e -t 5sec_updates
```

This copies data for transponder webId 83 and cartwheel accountId 18, from the given start-end time in millis since epoch.

We are providing a source (-b or --source) and destination (-g --target) via service account files for GCP.

Multiple tags are allowed with at least one being required.

### List

```bash
./replaystream list -b ./test-latinum-3cba82351b2d.json -t avl_status
```

This lists all tests with the provided tag from a source db.

### Tags

```bash
./replaystream tags --source test-latinum-3cba82351b2d.json
```

This lists all available tags in the source testing database.

### Replay

```bash
./replaystream replay ...
```
