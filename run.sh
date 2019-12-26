#!/bin/sh

case $1 in
	start)
		chmod +x ./bin/pushservice
		nohup ./bin/pushservice 2>&1 >> pushservice.log 2>&1 /dev/null &
		echo "服务已启动..."
		sleep 1
	;;
	stop)
		killall pushservice
		echo "服务已停止..."
		sleep 1
	;;
	restart)
		chmod +x ./bin/pushservice
		pid=$(ps x | grep -w "pushservice" | grep -v grep | awk '{print $1}')
		#echo $pid
		if [ ! "$pid" ];then
		chmod +x pushservice
		nohup ./bin/pushservice 2>&1 >> pushservice.log 2>&1 /dev/null &
		echo "服务已启动..."
		sleep 1
		else
		kill -2 $pid
		#echo "服务已重启..."
		#echo $pid
		sleep 1
		nohup ./bin/pushservice 2>&1 >> pushservice.log 2>&1 /dev/null &
		echo "服务已重启..."
		fi
		sleep 1
	;;
	*)
		echo "$0 {start|stop|restart}"
		exit 4
	;;
esac
