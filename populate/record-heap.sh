bigest=0
while true
do
        in_use=$(curl -S -s -X GET http://localhost:8081/debug/vars | tr "," "\n" | grep HeapInuse | awk -F: '{print $2}')
        echo $in_use, $bigest
  if [ "$in_use" -gt "$bigest" ]
        then
                echo RECORD
                bigest=$in_use
                curl -sS -X GET http://localhost:8081/debug/pprof/heap > ${bigest}.prof
        fi
        sleep 10
done