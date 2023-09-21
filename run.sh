#!/bin/bash

# 初始化变量
APP_NAME="appbff"
BUILD_CMD="go build -o $APP_NAME"
WATCH_PATH="./"
PORT=1005

# 定义 Ctrl+C 的处理函数
cleanup() {
  if [ "$old_pid" != "" ]; then
    kill $old_pid || true
  fi
  exit 0
}

# 捕获 Ctrl+C 信号，并调用 cleanup 函数进行处理
trap cleanup SIGINT

# 编译并启动程序
$BUILD_CMD && ./$APP_NAME &

# 获取初始进程ID
old_pid=$!

while true; do
  # 使用 fswatch 查找被更改的 .go 文件
  fswatch -1 -e ".*" -i "\\.go$" $WATCH_PATH

  echo "Detected changes, restarting..."

  # 杀死旧进程
  if [ "$old_pid" != "" ]; then
    kill $old_pid || true
  fi

  # 检查端口是否被占用，并如果需要的话杀掉占用进程
  port_pid=$(lsof -n -i4TCP:$PORT | grep LISTEN | awk '{print $2}' | uniq)
  if [ "$port_pid" != "" ]; then
    echo "Killing process $port_pid on port $PORT"
    kill -9 $port_pid || true
  fi

  # 重新编译并启动
  $BUILD_CMD && ./$APP_NAME &

  # 更新进程ID
  old_pid=$!
done
