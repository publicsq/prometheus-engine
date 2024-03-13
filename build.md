# Build

```
docker build --platform linux/amd64 -t us-docker.pkg.dev/psq-shd-operations/datasource-syncer/datasource-syncer:latest . -f cmd/datasource-syncer/Dockerfile
docker push us-docker.pkg.dev/psq-shd-operations/datasource-syncer/datasource-syncer:latest
```
