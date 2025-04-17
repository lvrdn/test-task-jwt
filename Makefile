app:
	docker-compose --profile disabled up -d

.PHONY: test
test:
	docker-compose --profile disabled -f test/init_db/docker-compose.yml up -d; 
	sleep 2;
	go test ./test;
	docker stop test_postgres;
	docker-compose -f test/init_db/docker-compose.yml -p init_db down