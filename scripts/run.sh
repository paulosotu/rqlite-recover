
#!/bin/ash

echo running: /bin/rqlite-recover -c -app $APP_NAME -l $LOG_LEVEL -t $UPDATE_INTERVAL_SECONDS -d $SHARED_STORAGE_FOLDER -n $NODE_HOSTNAME -http $RQLITE_HTTP_PORT -raft $RQLITE_RAFT_PORT -ltime $LIVENESS_TIMEOUT_SEC -rtime $READINESS_TIMEOUT_SEC -live $LIVENESS_FILENAME -ready $READINESS_FILENAME
/bin/rqlite-recover -c -app $APP_NAME -l $LOG_LEVEL -t $UPDATE_INTERVAL_SECONDS -d $SHARED_STORAGE_FOLDER -n $NODE_HOSTNAME -http $RQLITE_HTTP_PORT -raft $RQLITE_RAFT_PORT -ltime $LIVENESS_TIMEOUT_SEC -rtime $READINESS_TIMEOUT_SEC -live $LIVENESS_FILENAME -ready $READINESS_FILENAME