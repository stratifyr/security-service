#!/bin/bash

declare -a holidays=(
  "2025-01-26"
  "2025-03-17"
  "2025-04-14"
  "2025-04-18"
  "2025-05-01"
  "2025-08-15"
  "2025-10-02"
  "2025-10-23"
  "2025-10-31"
  "2025-12-25"
)

is_holiday() {
  local date_to_check=$1
  for holiday in "${holidays[@]}"; do
    if [[ "$holiday" == "$date_to_check" ]]; then
      return 0
    fi
  done
  return 1
}

start_date="2025-05-01"
end_date=$(date +%Y-%m-%d)

current_date="$start_date"
while [[ $(date -j -f "%Y-%m-%d" "$current_date" +%s) -le $(date -j -f "%Y-%m-%d" "$end_date" +%s) ]]; do
  day_of_week=$(date -j -f "%Y-%m-%d" "$current_date" +%u)

  if [[ "$day_of_week" -ge 6 ]]; then
    current_date=$(date -j -v+1d -f "%Y-%m-%d" "$current_date" +%Y-%m-%d)
    continue
  fi

  if is_holiday "$current_date"; then
    current_date=$(date -j -v+1d -f "%Y-%m-%d" "$current_date" +%Y-%m-%d)
    continue
  fi

  APP_ENV=test go run main.go run run -date="$current_date"

  sleep 2

  current_date=$(date -j -v+1d -f "%Y-%m-%d" "$current_date" +%Y-%m-%d)
done
