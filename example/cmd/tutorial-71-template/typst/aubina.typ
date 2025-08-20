
#let start_date = state("start_date", none)
#let current_date = state("current_date", none)
#let signed_date = state("signed_date", none)

#let weekdays = (
  "Montag",
  "Dienstag",
  "Mittwoch",
  "Donnerstag",
  "Freitag",
  "Samstag",
  "Sonntag",
)


#let months = (
  "Januar",
  "Februar",
  "März",
  "April",
  "Mai",
  "Juni",
  "Juli",
  "August",
  "September",
  "October",
  "November",
  "Dezember",
)


#let aubina_title(
  title: none,
  subtitle: none,
  author: (),
  trainer: (),
  training: (),
  company: (),
  abstract: [],
  doc,
) = {
  set page(
    paper: "a4",
  )

  set text(
    font: "Arial",
    size: 12pt,
    lang: "de",
  )

  v(2cm)

  set align(center)
  text(25pt, title)

  set align(center)
  text(20pt, subtitle)

  v(1cm)
  par[
    *#author.name* \
    geb. #author.birthdate.display("[day].[month].[year]") \
    #author.address.city \
    #author.address.street
  ]
  par[
    *Ausbildung* \
    #training.start.display("[day].[month].[year]")
    \-
    #training.end.display("[day].[month].[year]")
  ]

  v(2cm)

  image("logo.png", width: 30%)

  v(3cm)


  par[
    *Ausbildungsbetrieb* \
    #company.name \
    #company.address.city \
    #company.address.street \
  ]
  grid(columns: 1)[
    *Ausbilder* \
    #trainer.name \
  ]

  pagebreak()

  set align(left)

  start_date.update(training.start)

  set page(
    margin: (x: 1.8cm, top: 4cm, bottom: 7cm),
    paper: "a4",
    header: align(right)[
      #image("logo.png", width: 8%)
    ],
    footer: context align(left)[

      #table(
        stroke: none,
        columns: (1fr, 2fr),

        [
          #table(
            stroke: none,
            columns: (1fr, 1fr),
            [
              #let d = signed_date.get()
              #d.display("[day].[month].[year]")
              #set align(center)
              #line(length: 150%)
              Datum
            ],
          )
        ],

        [
          #table(
            stroke: none,
            columns: (1fr, 1fr),
            [
              \
              #set align(center)
              #line(length: 75%)
              Auszubildende/-r
            ],
            [
              \
              #set align(center)
              #line(length: 75%)
              Ausbilder/-in
            ],
          )
          #v(0.8cm)
        ],

        [
          #let y = 1 + calc.floor((current_date.get() - start_date.get()).days() / 365)
          #set text(fill: rgb("#6d6d6d"))
          Ausbildungsjahr: #y
        ],
        [
          #table(
            stroke: none,
            columns: (auto, 1fr),
            [
              #set align(left)
              #set text(fill: rgb("#6d6d6d"))
              Ausbildungsnachweis | #author.name
            ],
            [
              #set align(right)
              #counter(page).display("1")
            ],
          )
        ],
      )
    ],
  )

  doc
}

#let aubina_day(
  date: (),
  // duration: none,
  // description: none,
  place: "Betrieb",
  tasks: (),
  doc,
) = {
  table(
    stroke: none,
    gutter: 3pt,
    columns: (auto, 1fr, auto),

    [
      #text(25pt, fill: rgb("#1a843b"), str(date.day()))
    ],
    [

      #table(
        stroke: none,
        columns: (auto, 1fr),
        [
          #text(fill: rgb("#1a843b"), months.at(int(date.display("[month]")) - 1)) \
          #weekdays.at(
            int(date.display(
              "[weekday repr:monday]",
            ))
              - 1,
          )
        ],
        [
          Ort\
          #place
        ],
      )

    ],
    [],
    table.hline(start: 0),

    [],
  )
  v(-10pt)

  table(
    stroke: none,
    gutter: 5pt,
    columns: (1fr, 3cm),
    ..tasks
      .map(task => {
        (
          [
            #task.description
          ],
          [
            #if str(calc.floor(task.duration.minutes() / 60)).len() == 1 {
              "0"
            }#str(calc.floor(task.duration.minutes() / 60))h
            #if str(calc.rem(task.duration.minutes(), 60)).len() == 1 {
              "0"
            }#str(calc.rem(task.duration.minutes(), 60))m
          ],
        )
      })
      .flatten()
  )

  context current_date.update(date)

  doc
}



