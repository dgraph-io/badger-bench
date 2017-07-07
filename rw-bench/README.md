To run: go test -v 

=== RUN   TestPutAndIterate
Value size: 1024
Num unique keys: 2855231
rocks iteration time:  1.717578587s
rocks time:  13.540254938s
Num unique keys: 2855231
badger iteration time:  1.231884332s
badger time:  16.91380382s

=== RUN   TestPutAndIterate
Value size: 128
Num unique keys: 2855316
rocks iteration time:  1.641643682s
rocks time:  9.72966573s
Num unique keys: 2855316
badger iteration time:  1.207451184s
badger time:  13.3816506s

=== RUN   TestPutAndIterate
Value size: 0
Num unique keys: 2855279
rocks iteration time:  1.652722679s
rocks time:  8.782233328s
Num unique keys: 2855279
badger iteration time:  2.109307851s
badger time:  15.168629622s
