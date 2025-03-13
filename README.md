# Корпаротивный портал KamaDiesel backend
Краткое описание проекта

## Содержание
- [Технологии](#технологии)
- [Разработка](#разработка)
- [Git Flow](#памятка-по-git-flow)
- [Deploy](#deploy)
- [ЧаВо](#faq)
- [Источники](#источники)

## Технологии
- [Golang](URL)
- [PostgreSQL](URL)


## Разработка

### Требования
Для установки и запуска проекта, необходим [golang](url) vN+

## Памятка по Git Flow
### Основные ветки

1. **Main** — основная стабильная ветка, содержащая код, готовый к выпуску.
2. **Develop** — основная ветка разработки, в которую интегрируются изменения перед выпуском.

Кроме этих двух, создаются временные ветки для разработки функций, подготовки релиза и экстренных исправлений.

### Типы временных веток

1. **Feature** — для разработки новой функции.
2. **Release** — для подготовки к выпуску.
3. **Hotfix** — для экстренных исправлений.

### Процесс работы

#### 1. Разработка новой функции: Ветка `Feature`

1. Переключитесься на ветку `develop`:

    ```bash
    git checkout develop
    ```

2. Создайть новую ветку для фичи, например `feature/<feature-name>`:

    ```bash
    git checkout -b feature/<feature-name>
    ```

3. Разработать функцию в ветке `feature/<feature-name>`. После завершения работы выполнить коммит всех изменений:

    ```bash
    git add .
    git commit -m "Feature: <feature-name>"
    ```

4. Влить изменения из ветки фичи обратно в `develop`:

    ```bash
    git checkout develop
    git merge feature/<feature-name>
    ```

5. Удалить ветку фичи:

    ```bash
    git branch -d feature/<feature-name>
    ```

#### 2. Подготовка к релизу: Ветка `Release`

1. Переключиться на ветку `develop`:

    ```bash
    git checkout develop
    ```

2. Создайть ветку `release/<release-version>` от `develop`, где `<release-version>` — номер версии, например, `1.0.0`:

    ```bash
    git checkout -b release/<release-version>
    ```

3. Внести все необходимые исправления и протестировать в ветке `release/<release-version>`. После завершения внести изменения и выполнить коммит:

    ```bash
    git add .
    git commit -m "Prerelease: <release-version>"
    ```

4. Переключиться на `main` и влить изменения из `release/<release-version>`:

    ```bash
    git checkout main
    git merge release/<release-version>
    ```

5. Создать тег для релиза:

    ```bash
    git tag -a <release-version> -m "Release <release-version>"
    ```

6. Переключиться обратно на `develop` и влить в него изменения из `release/<release-version>`:

    ```bash
    git checkout develop
    git merge release/<release-version>
    ```

7. Удалить ветку `release`:

    ```bash
    git branch -d release/<release-version>
    ```

#### 3. Экстренное исправление: Ветка `Hotfix`

1. Переключитесь на `main`:

    ```bash
    git checkout main
    ```

2. Создайте ветку `hotfix/<hotfix-version>` от `main`, где `<hotfix-version>` — номер версии хотфикса, например, `1.0.1`:

    ```bash
    git checkout -b hotfix/<hotfix-version>
    ```

3. Внести необходимые исправления и выполнить коммит:

    ```bash
    git add .
    git commit -m "Hotfix: <hotfix-version>"
    ```

4. Влить изменения в `main`:

    ```bash
    git checkout main
    git merge hotfix/<hotfix-version>
    ```

5. Создать тег для хотфикса:

    ```bash
    git tag -a <hotfix-version> -m "HotFix <hotfix-version>"
    ```

6. Переключиться на `develop` и влить в него изменения из `hotfix/<hotfix-version>`:

    ```bash
    git checkout develop
    git merge hotfix/<hotfix-version>
    ```

7. Удалить ветку хотфикса:

    ```bash
    git branch -d hotfix/<hotfix-version>
    ```
### Основные принципы версионирования
#### **Семантическое версионирование (Semantic Versioning)**: формата `MAJOR.MINOR.PATCH`, например, `1.2.3`.

- **MAJOR** — крупный релиз, в котором могут быть изменения, несовместимые с предыдущими версиями. Увеличивается, если добавлены большие изменения или новая функциональность.
- **MINOR** — минорный релиз, добавляющий новые функции, но сохраняющий совместимость с предыдущими версиями.
- **PATCH** — исправление ошибок или мелкие улучшения, не добавляющие нового функционала.

#### **Теги для версий**:

   При завершении работы над ветками `release` или `hotfix` создается тег, совпадающий с номером версии, например, `1.0.0` или `1.2.1`.

## Deploy
...

## FAQ

### Зачем вы разработали этот проект?
Чтобы был.

## Источники
...