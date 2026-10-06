TEST_PATHS = ./internal/repository/... ./internal/usecase/... 
CONFIG_FILE = ./config/config.yaml
ALL_PATHS = $(TEST_PATHS)

COVER_OUT = coverage.out
COVER_HTML = coverage.html

ALLURE_RESULTS = allure-results
ALLURE_REPORT  = allure-report
JUNIT_REPORT   = junit-report.xml

SONAR_PROJECT_PROPS = sonar-project.properties
SONAR_SCANNER       = sonar-scanner
SONAR_HOST_URL ?= http://localhost:9100


GREEN  = \033[0;32m
CYAN   = \033[0;36m
YELLOW = \033[1;33m
RED    = \033[0;31m
RESET  = \033[0m

.PHONY: test test-full test-shuffle coverage-cli coverage-html clean help \
        test-restore test-allure allure-open clean-allure \
        coverage-dashboard coverage-dashboard-open \
        sonar sonar-setup sonar-check

COVERAGE_DASHBOARD = coverage-dashboard.html

.PHONY: coverage-dashboard coverage-dashboard-open

coverage-dashboard: $(COVER_OUT)
	@echo "$(CYAN) Генерация веб-дашборда покрытия по модулям (gocov)...$(RESET)"
	@gocov convert $(COVER_OUT) | gocov-html > $(COVERAGE_DASHBOARD)
	@echo "$(GREEN)  Дашборд успешно сохранен в $(COVERAGE_DASHBOARD)$(RESET)"

coverage-dashboard-open:
	@echo "$(CYAN) Дашборд покрытия в браузере...$(RESET)"
	@open $(COVERAGE_DASHBOARD) || xdg-open $(COVERAGE_DASHBOARD) || start $(COVERAGE_DASHBOARD)

prism:
	docker run --rm -v ${PWD}:/tmp -p 4010:4010 stoplight/prism:4 mock -h 0.0.0.0 /tmp/api/openapi.yaml

yaml-lint:
	docker run --rm -v ${PWD}:/tmp stoplight/spectral lint /tmp/api/openapi.yaml --ruleset /tmp/.spectral.yaml

test:
	@echo "$(CYAN)Запуск тестов с расчетом покрытия...$(RESET)"
	@go test -p 1 -v -tags=postgres -coverprofile=$(COVER_OUT) -covermode=count $(ALL_PATHS)
	@echo "\n$(GREEN)Итоговое покрытие по функциям (CLI):$(RESET)"
	@go tool cover -func=$(COVER_OUT) | tail -n 1
	@echo "$(GREEN) Полный отчет сохранен в $(COVER_OUT).$(RESET)"

test-full:
	@echo "$(CYAN)[1/5] Запуск тестов на текущей конфигурации (Postgres + Redis)...$(RESET)"
	@go test -p 1 -v -tags=postgres -shuffle=on -count=1 -coverprofile=cover-postgres.out -covermode=count $(ALL_PATHS)
	
	@echo "$(CYAN)[2/5] Переключение БД на Cassandra + Tarantool...$(RESET)"
	@sed -i '' 's/primary_type: "postgres"/primary_type: "cassandra"/' $(CONFIG_FILE)
	@sed -i '' 's/cache_type: "redis"/cache_type: "tarantool"/' $(CONFIG_FILE)
	@echo "$(GREEN)  Активные типы БД сейчас:$(RESET)"
	@grep -E "primary_type|cache_type" $(CONFIG_FILE) | head -n 2
	
	@echo "$(CYAN)[3/5] Запуск тестов (shuffle + coverage) на новой БД...$(RESET)"
	@go test -p 1 -v -tags=cassandra -shuffle=on -count=1 -coverprofile=cover-cassandra.out -covermode=count $(ALL_PATHS) 

	@echo "$(CYAN)[4/5] Восстановление конфигурации на Postgres + Redis...$(RESET)"
	@make test-restore
	
	@echo "$(CYAN)[5/5] Объединение отчетов о покрытии (gocovmerge)...$(RESET)"
	@which gocovmerge > /dev/null || (echo "Устанавливаю gocovmerge..." && go install github.com/wadey/gocovmerge@latest)
	@gocovmerge cover-postgres.out cover-cassandra.out > $(COVER_OUT)
	@rm -f cover-postgres.out cover-cassandra.out
	
	@echo "$(GREEN)Итоговое покрытие проекта:$(RESET)"
	@go tool cover -func=$(COVER_OUT) | tail -n 1
	@go tool cover -html=$(COVER_OUT) -o $(COVER_HTML)
	@echo "$(GREEN)HTML-отчет обновлен ($(COVER_HTML)).$(RESET)"

test-shuffle:
	@echo "$(CYAN)Запуск тестов в случайном порядке (-shuffle=on)...$(RESET)"
	@go test -p 1 -v -tags=postgres -shuffle=on -count=1 -coverprofile=$(COVER_OUT) -covermode=count $(ALL_PATHS)
	@echo "\n$(GREEN)Итоговое покрытие (CLI):$(RESET)"
	@go tool cover -func=$(COVER_OUT) | tail -n 1

coverage-cli: $(COVER_OUT)
	@echo "$(CYAN)Детальный отчет по покрытию (CLI):$(RESET)"
	@go tool cover -func=$(COVER_OUT)

coverage-html: $(COVER_OUT)
	@echo "$(CYAN) Генерация HTML отчета...$(RESET)"
	@go tool cover -html=$(COVER_OUT) -o $(COVER_HTML)
	@echo "$(GREEN)Отчет успешно сохранен в $(COVER_HTML)$(RESET)"
	@open $(COVER_HTML) || xdg-open $(COVER_HTML) || start $(COVER_HTML)

$(COVER_OUT):
	@go test -coverprofile=$(COVER_OUT) -covermode=count $(ALL_PATHS)

test-restore:
	@sed -i '' 's/primary_type: "cassandra"/primary_type: "postgres"/' $(CONFIG_FILE)
	@sed -i '' 's/cache_type: "tarantool"/cache_type: "redis"/' $(CONFIG_FILE)
	@echo "$(GREEN) Конфиг восстановлен на Postgres + Redis.$(RESET)"


test-allure: clean-allure
	@echo "$(CYAN)Запуск тестов с генерацией Allure отчёта...$(RESET)"
	
	@go test -p 1 -v -tags=postgres -shuffle=on -count=1 -coverprofile=$(COVER_OUT) -covermode=count $(ALL_PATHS) > /tmp/go-test-output.log 2>&1 || true
	
	@go tool cover -html=$(COVER_OUT) -o coverage.html
	
	@which go-junit-report > /dev/null || (echo "$(YELLOW) Устанавливаю go-junit-report...$(RESET)" && go install github.com/jstemmer/go-junit-report/v2@latest)
	@cat /tmp/go-test-output.log | go-junit-report > $(JUNIT_REPORT)
	
	@mkdir -p $(ALLURE_RESULTS)
	@cp $(JUNIT_REPORT) $(ALLURE_RESULTS)/
	@echo "Go.Version=$(shell go version | awk '{print $$3}')" > $(ALLURE_RESULTS)/environment.properties
	@echo "Coverage=$(shell go tool cover -func=$(COVER_OUT) | tail -n 1 | awk '{print $$NF}')" >> $(ALLURE_RESULTS)/environment.properties
	@echo "Detailed Coverage Report=<a href='../coverage.html' target='_blank'>Открыть HTML отчет покрытия</a>" >> $(ALLURE_RESULTS)/environment.properties
	
	@allure generate $(ALLURE_RESULTS) -o $(ALLURE_REPORT) --clean
	
	@echo "$(GREEN) Allure отчёт сгенерирован в $(ALLURE_REPORT)/$(RESET)"
	@echo "$(GREEN) HTML отчет покрытия сгенерирован в coverage.html$(RESET)"

allure-open:
	@echo "$(CYAN) Allure отчёт...$(RESET)"
	@allure open $(ALLURE_REPORT)

clean-allure:
	@rm -rf $(ALLURE_RESULTS) $(ALLURE_REPORT) $(JUNIT_REPORT)
	@echo "$(GREEN) Allure результаты очищены$(RESET)"

clean: clean-allure
	@rm -f $(COVER_OUT) $(COVER_HTML) cover-postgres.out cover-cassandra.out
	@rm -rf .scannerwork
	@echo "$(GREEN) Файлы покрытия и артефакты Sonar удалены.$(RESET)"


help:
	@echo "$(CYAN)Доступные команды:$(RESET)"
	@echo "  make test          - Запуск тестов repository + usecase + покрытие в CLI"
	@echo "  make test-full     - Полный прогон: Postgres -> Cassandra -> объединение coverage (gocovmerge)"
	@echo "  make test-shuffle  - Запуск тестов в случайном порядке"
	@echo "  make coverage-cli  - Подробный отчет по покрытию в терминале"
	@echo "  make coverage-html - Сгенерировать и открыть HTML-отчет"
	@echo "  $(YELLOW)make test-allure$(RESET)  - $(GREEN) Запустить тесты и сгенерировать Allure отчёт$(RESET)"
	@echo "  $(YELLOW)make sonar$(RESET)          - Прогон тестов + отправка в SonarQube ($(SONAR_HOST_URL))"
	@echo "  $(YELLOW)make sonar-cloud$(RESET)    - Прогон тестов + отправка в SonarCloud (нужен SONAR_TOKEN)"
	@echo "  make sonar-setup    - Установить sonar-scanner (macOS, через brew)"
	@echo "  make sonar-check    - Проверить наличие sonar-scanner и конфига"
	

sonar-check:
	@which $(SONAR_SCANNER) > /dev/null || \
	  (echo "$(RED) sonar-scanner не найден. Установите его:$(RESET)" && \
	   echo "  brew install sonar-scanner           # macOS" && \
	   echo "  или скачайте с https://docs.sonarqube.org/latest/analyzing-source-code/scanners/sonarscanner/" && \
	   exit 1)
	@test -f $(SONAR_PROJECT_PROPS) || \
	  (echo "$(RED) $(SONAR_PROJECT_PROPS) не найден в корне проекта$(RESET)" && exit 1)

sonar-setup:
	@echo "$(CYAN)Установка sonar-scanner...$(RESET)"
	@which brew > /dev/null || (echo "$(RED) Homebrew не найден$(RESET)" && exit 1)
	@brew install sonar-scanner
	@echo "$(GREEN)  sonar-scanner установлен$(RESET)"

sonar: sonar-check
	@echo "$(CYAN)[1/2] Прогон тестов с покрытием...$(RESET)"
	@go test -p 1 -tags=postgres -count=1 \
	  -coverprofile=$(COVER_OUT) -covermode=count \
	  $(ALL_PATHS)
	@echo "\n$(GREEN)Покрытие:$(RESET)"
	@go tool cover -func=$(COVER_OUT) | tail -n 1

	@echo "$(CYAN)[2/2] Отправка в SonarQube ($(SONAR_HOST_URL))...$(RESET)"
	@$(SONAR_SCANNER) \
	  -Dsonar.host.url=$(SONAR_HOST_URL) \
	  -Dsonar.projectBaseDir=. \
	  -Dsonar.go.coverage.reportPaths=$(COVER_OUT)
	@echo "$(GREEN) Анализ завершён. Отчёт: $(SONAR_HOST_URL)/dashboard?id=Neratus_geoguide$(RESET)"

test-docker:
	@echo "$(CYAN)Запуск тестового стенда в Docker...$(RESET)"
	@docker compose -f docker-compose.test.yml up --build --abort-on-container-exit --exit-code-from tester
	@echo "$(GREEN)Тесты завершены. Отчеты сохранены в ./allure-results и ./coverage$(RESET)"

clean-docker:
	@echo "$(YELLOW)Очистка тестового Docker-окружения...$(RESET)"
	@docker compose -f docker-compose.test.yml down -v
	@rm -rf ./allure-results ./coverage
	@echo "$(GREEN)Окружение очищено$(RESET)"

test-ci:
	@echo "$(CYAN)Запуск CI тестового стенда (App + Postgres + Tester)...$(RESET)"
	@docker compose -f docker-compose.ci.yml up \
	    --build --abort-on-container-exit --exit-code-from tester; \
	STATUS=$$?; \
	echo "$(YELLOW)Очистка CI Docker-окружения...$(RESET)"; \
	docker compose -f docker-compose.ci.yml down -v --remove-orphans; \
	exit $$STATUS

clean-ci:
	@echo "$(YELLOW)Очистка CI Docker-окружения...$(RESET)"
	@docker compose -f docker-compose.ci.yml down -v
	@rm -rf ./allure-results ./coverage
	@echo "$(GREEN)Окружение очищено$(RESET)"