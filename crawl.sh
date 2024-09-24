#!/bin/bash
# 指定最忙的前N个线程并打印堆栈
max_thread_count=50
# 指定cpu占比统计的采样间隔，单位为毫秒
interval=2000
# arthas路径
arthas_path="/tmp/arthas"
# arthas的下载地址
arthas_boot_download_url="https://arthas.aliyun.com/arthas-boot.jar"

# jps命令不存在代表非java应用或没有jdk相关工具
if [ ! "$(type jps)" ]; then
  exit 0
fi

cd ~/ || exit

# arthas-boot.jar 不存在则下载
if [ ! -f "${arthas_path}/arthas-boot.jar" ]; then
  mkdir -p /tmp/arthas
  curl -sO "$arthas_boot_download_url" && mv arthas-boot.jar ${arthas_path}/
fi

unset JAVA_TOOL_OPTIONS
# 处理pid 1非java应用的情况
pid=$(jps -l | grep -v "Jps" | grep -v "arthas-boot.jar" | grep -v "process information unavailable" | awk '{print $1}')

# 获取繁忙的前50个线程
java -jar ${arthas_path}/arthas-boot.jar "${pid}" -c "thread -n ${max_thread_count} -i ${interval}"
