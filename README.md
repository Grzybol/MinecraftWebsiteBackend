---
# website_backend

Backend projektu napisany w języku Go.

## Spis treści
- [Opis projektu](#opis-projektu)
- [Wymagania](#wymagania)
- [Instalacja](#instalacja)
- [Uruchamianie](#uruchamianie)
- [Struktura projektu](#struktura-projektu)
- [Konfiguracja](#konfiguracja)
- [Testowanie](#testowanie)
- [Licencja](#licencja)

## Opis projektu

Projekt **website_backend** to serwerowa aplikacja napisana w Go, która stanowi backend dla strony internetowej. Odpowiada za obsługę logiki biznesowej, przechowywanie danych oraz komunikację z frontendem.

## Wymagania

- Go (wersja 1.18 lub wyższa)
- (Opcjonalnie) Docker
- Inne zależności opisane w pliku `go.mod`

## Instalacja

1. Sklonuj repozytorium:
   ```bash
   git clone https://github.com/Grzybol/website_backend.git
   cd website_backend
   ```

2. Pobierz zależności:
   ```bash
   go mod download
   ```

## Uruchamianie

Aby uruchomić serwer lokalnie:
```bash
go run main.go
```
Lub zbudować binarkę:
```bash
go build -o backend
./backend
```

### Docker

Uruchomienie aplikacji w kontenerze (z automatycznym restartem po błędach oraz cyklicznie co 24h):
```bash
docker compose up --build
```

Domyślny interwał restartu jest kontrolowany przez zmienną `RESTART_INTERVAL_SECONDS` i wynosi 86400 sekund (24h). Możesz go zmienić w `docker-compose.yml` lub przez nadpisanie zmiennej środowiskowej.

## Struktura projektu

Krótki opis głównych plików/katalogów:
- `main.go` – punkt wejścia aplikacji
- `internal/` – logika aplikacji (np. obsługa API, modele, serwisy)
- `config/` – pliki konfiguracyjne
- `go.mod`, `go.sum` – zależności projektu

_(Dostosuj powyższe według rzeczywistej struktury projektu)_

## Konfiguracja

Zmienne konfiguracyjne (np. port, połączenie z bazą danych) ustawiane są w pliku `.env` lub bezpośrednio w kodzie. Przykład pliku `.env`:
```
PORT=8080
DATABASE_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable
```

## Testowanie

Aby uruchomić testy:
```bash
go test ./...
```

## Licencja

Projekt dostępny na licencji MIT.

---

Jeśli chcesz, mogę dostosować ten plik do bardziej szczegółowych informacji o funkcjonalnościach lub strukturze kodu – daj znać, jeśli potrzebujesz!
