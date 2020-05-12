# point Tilt at the existing docker-compose configuration.
docker_compose("./docker-compose.yml")
docker_build('server', '.', dockerfile='Dockerfile.server',
  live_update = [
    sync('.', '/cms'),
    run('go build -o server', trigger=''),
    restart_container()
  ])