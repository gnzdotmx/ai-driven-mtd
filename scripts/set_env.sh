
export SELECTED_PORT="$1"
export SELECTED_FORMAT="$2"
export SELECTED_LANGUAGE="$3"
export SELECTED_OS="$4"
LANGUAGE=$3
OS=$4

docker compose -f ./docker/docker-compose.yml down 2>/dev/null
case $OS in
  golang)
    case $LANGUAGE in
      golang)
        docker compose -f ./docker/docker-compose.yml up -d app_golang_golang 2>/dev/null
        ;;
      python)
        docker compose -f ./docker/docker-compose.yml up -d app_golang_python 2>/dev/null
        ;;
      *)
        docker compose -f ./docker/docker-compose.yml up -d app_golang_golang 2>/dev/null
        ;;
    esac
    ;;
  python)
    case $LANGUAGE in
      python)
        docker compose -f ./docker/docker-compose.yml up -d app_python_python 2>/dev/null
        ;;
      golang)
        docker compose -f ./docker/docker-compose.yml up -d app_python_golang 2>/dev/null
        ;;
      *)
        docker compose -f ./docker/docker-compose.yml up -d app_python_python 2>/dev/null
        ;;
    esac
    ;;
  ubuntu)
    case $LANGUAGE in
      golang)
         docker compose -f ./docker/docker-compose.yml up -d app_ubuntu_golang 2>/dev/null
         ;;
      python)
         docker compose -f ./docker/docker-compose.yml up -d app_ubuntu_python 2>/dev/null
         ;;
      *)
         docker compose -f ./docker/docker-compose.yml up -d app_ubuntu_python 2>/dev/null
         ;;
    esac
    ;;
  *)
    echo "Unknown OS: $OS"
    exit 1
    ;;
esac
