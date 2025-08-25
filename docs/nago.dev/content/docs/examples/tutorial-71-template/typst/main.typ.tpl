#import "aubina.typ": aubina_day, aubina_title, signed_date, start_date
#show: aubina_title.with(
  title: [
    {{.Title}}
  ],
  subtitle: [
    {{.Subtitle}}
  ],
  author: (
    name: "{{.Author.Name}}",
    // birthdate: "{{.Author.Birthday}}",
    birthdate: datetime(
      year: {{.Author.Birthday.Year}},
      month: {{.Author.Birthday.Month}},
      day: {{.Author.Birthday.Day}},
    ),
    address: (
      street: "{{.Author.Address.Street}}",
      city: "{{.Author.Address.City}}",
    ),
  ),
  company: (
    name: "{{.Company.Name}}",
    address: (
      street: "{{.Company.Address.Street}}",
      city: "{{.Company.Address.City}}",
    ),
  ),
  trainer: (
    name: "{{.Trainer}}",
  ),
  training: (
    start: datetime(
      year: {{.Training.Start.Year}},
      month: {{.Training.Start.Month}},
      day: {{.Training.Start.Day}},
    ),
    end: datetime(
      year: {{.Training.End.Year}},
      month: {{.Training.End.Month}},
      day: {{.Training.End.Day}},
    ),
  ),
)

{{range $i, $entry := .Entries}}
  {{if $entry.Kind}}
    #signed_date.update(datetime(
      day: {{$entry.Date.Day}},
      month: {{$entry.Date.Month}},
      year: {{$entry.Date.Year}})
    )
  {{else}}
    #show: aubina_day.with(
      date: datetime(
        year: {{$entry.Date.Year}},
        month: {{$entry.Date.Month}},
        day: {{$entry.Date.Day}},
      ),
      tasks: (
        {{range $i, $task := $entry.Tasks}}
          (
            description: "{{$task.Description}}",
            duration: duration(minutes: {{$task.DurationInMinutes}}),
          ),
        {{end}}
      ),
      place: "{{$entry.Place}}",
    )
  {{end}}
{{end}}


//----

