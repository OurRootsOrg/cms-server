# point Tilt at the existing docker-compose configuration.
docker_compose("./docker-compose.yml")
docker_build('server', '.', dockerfile='Dockerfile.server',
  live_update = [
    sync('.', '/cms'),
    run('go generate && go build -o server', trigger=''),
    restart_container()
  ])
docker_build('recordswriter', '.', dockerfile='Dockerfile.recordswriter',
  live_update = [
    sync('.', '/cms'),
    run('go generate && go build -o server', trigger=''),
    restart_container()
    ])
docker_build('search', '.', dockerfile='Dockerfile.search',
  live_update = [
    sync('.', '/cms'),
    run('go generate && go build -o search', trigger=''),
    restart_container()
  ])