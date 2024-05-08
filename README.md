# GCP IAM dumper

`gcp-iam-dumper` is a CLI tool designed to facilitate the management of IAM (Identity and Access Management) data within Google Cloud Platform (GCP). It allows users to dump IAM data into a SQLite database, export the database to CSV format, and upload files to Google Cloud Storage (GCS).

## Features

- **Dump IAM Data**: Collect IAM data from a GCP organization and save it into a SQLite database.
- **Export to CSV**: Convert the SQLite database into CSV format for easy sharing and analysis.
- **Upload to GCS**: Easily upload any file, including the exported CSV, to a specified GCS bucket.

## Usage

### General Syntax

```bash
gcp-iam-dumper [command]
```

Available Commands:
- `completion`: Generate the autocompletion script for the specified shell.
- `dump`: Dump IAM data into a SQLite database.
- `export`: Export SQLite database to CSV.
- `help`: Display help information about any command.
- `upload`: Upload files to GCS.

### Dumping IAM Data

To dump IAM data into a SQLite database, use:

```bash
gcp-iam-dumper dump --gcpOrgId <org_id> --quotaProjectId <project_id> --workspaceOrgId <workspace_org_id> [--sqliteFile <path/to/database.db>]
```

- `--gcpOrgId`: GCP organization ID (mandatory).
- `--quotaProjectId`: The quota project ID used for Directory API/Cloud Identity API (mandatory).
- `--workspaceOrgId`: Workspace organization ID (mandatory).
- `--sqliteFile`: Path to the SQLite file (optional, default "./database.db").

### Exporting to CSV

To export the SQLite database to CSV:

```bash
gcp-iam-dumper export [--exportDir <path/to/export>] [--sqliteFile <path/to/database.db>]
```

- `--exportDir`: Directory used to dump CSV exports (optional, default "./export").
- `--sqliteFile`: Path to the SQLite file (optional, default "./database.db").

### Uploading to GCS

To upload files to GCS:

```bash
gcp-iam-dumper upload --bucketName <bucket_name> [--srcPath <path/to/source>]
```

- `--bucketName`: GCS Bucket name where files are uploaded (mandatory).
- `--srcPath`: Path to upload, can be a file or a directory (non-recursive) (optional, default "./export").

## Example queries

### Lists individual permissions assignments
```
select *
from resource_role_principal rrp
join principal p on p.name=rrp.principal_name
where p.type='user';

```

### Recursively lists members of a principal
```
WITH RECURSIVE child_principals(id, name, type) AS (
    SELECT id, name, type
    FROM principal
    WHERE name = 'my-team@example.com'
    UNION ALL
    SELECT p.id, p.name, p.type
    FROM principal p
             INNER JOIN principal_hierarchy ph ON p.id = ph.child_id
             INNER JOIN child_principals cp ON ph.parent_id = cp.id
)
SELECT id, name, type FROM child_principals;
```

### Recursively lists parents of a principal
```
WITH RECURSIVE parent_principals(id, name, type) AS (
    SELECT id, name, type
    FROM principal
    WHERE name = 'my-team@example.com'
    UNION ALL
    SELECT p.id, p.name, p.type
    FROM principal p
             INNER JOIN principal_hierarchy ph ON p.id = ph.parent_id
             INNER JOIN parent_principals pp ON ph.child_id = pp.id
)
SELECT id, name, type FROM parent_principals;
```

### Lists everything a principal can do
```
WITH RECURSIVE parent_principals(id, name, type) AS (
    SELECT id, name, type
    FROM principal
    WHERE name = 'my-team@example.com'
    UNION ALL
    SELECT p.id, p.name, p.type
    FROM principal p
             INNER JOIN principal_hierarchy ph ON p.id = ph.parent_id
             INNER JOIN parent_principals pp ON ph.child_id = pp.id
)
select
    rrp.principal_name,
    h.name,
    rrp.role_id,
    rrp.conditional,
    rrp.asset_type,
    rrp.resource_id
from resource_role_principal rrp
join parent_principals pp on pp.name = rrp.principal_name
join hierarchy h on h.id = rrp.hierarchy_id;
```


### Lists deleted principals permissions to be cleaned up
```
select *
from resource_role_principal
where principal_name like '%?uid=%';
```

### Lists external users
```
select
    h.name,
    rrp.*
from resource_role_principal rrp
join hierarchy h on h.id=rrp.hierarchy_id
where rrp.principal_name not like '%@example.com%'
and   rrp.principal_name not like '%gserviceaccount.com%'
and   rrp.principal_name not like '%[%'
and   rrp.principal_name != 'allUsers';
```

## Building

```
go build -o gcp-iam-dumper cmd/main.go
```
