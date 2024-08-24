
### Info 

This is a fork of the 'zeit' repo which can be found [here](https://github.com/mrusme/zeit/)

Much thanks to Mrusme for his work and for making it available to the public.   
I have decided to fork it as I wanted to have some additions and be able to use this accorss different PC's.  
Sadly my setup even with changes did not work with the requiered '.editorconfig',so I couldn't contribute upstream. 

> **[Jump to the original project](https://github.com/mrusme/zeit)**

--- 
---

Table of contents: 

- [What?](#What-is-this)
- [Changes to know](#changes-to-know)
- [How to build/install](#building)
- [What is new](#New-additions)
- [Issues](#Issues)

## What is this?

This project is a `cli` that allows you to track how much time you spend on a given task on a given project.
You can export the statistics to a CSV, display them in the Terminal via a nice statistics screen, or just list them out.

<insert picture of good screenshot>

## Changes to know

I just realeased the `v0.2.0`, you can head over there to read about the changes too. 
[Read the release](https://github.com/MaximilianSoerenPollak/zeit/releases/tag/v.0.2.0)

Here are some changes I have made to my fork, that you should be aware of if you use this one.  

#### V0.2.0 Changes 
- Change to SQLITE3 Database 
- Remove 'tyme' completely (from import & export)
- Removed some commands to streamline the Process 
- Added 'colors' to the 'stats' page. 
- Removed Git/hub linking feature 

#### V0.1.0 Changes
- Removed 'extras' as I had no need for them 
- Changed Go Version to 1.22.3 (to make use of new Bugfixes etc.)
- Implemented standard 'GoFmt' formatting 
- Changed 'default' arguments for exporting and added options to it.


### Building
You can build the project by using the 'makefile' or by just using go itself. 
```sh
make VERSION=0.2.0 # to make the newest version.
```
or via go 
```sh 
go build zeit.go  #This however makes it so the 'version' command is not set 
```
After you just have to move the `zeit` binary to into your 'PATH' in order to just use it as `zeit` from anywhere.  
You can accomplish this by moving the binary to `/bin/`.   
If you rather would just the user have execute this binary you can also make a `bin` folder in your home directory and add the zeit binary in there and add the folder to path with adding this to the *.bashrc* or your shells config file  
`export PATH="$PATH:/usr/local/bin"`

###  New additions


#### Export to CSV 
I have added the option to allow `zeit` to also export to 'csv'. Currently it will use `;` as a seperator
but the plan is to have that configurable in the future.  
How to use it: 

```sh 
$ zeit export --format "csv" --file-name "exportedTimes.csv"
```

#### Extra options to export 
There are some new additions to the export command.
- `date` in a *YYYY-MM-DD Format*. Flag -> `--date` 
- `hours` which just adds the calulated hours per task to the export. Flag -> `--hours-decimal` 

They can be accessed via flags (booleans) to turn them on/off. By default both are `true` so your exports will contain them
If you'd like either, or both to be false just supply the correct flag with `false`, like so:  
```sh 
$ zeit export --hours-decimal false --date false
```

- `exportAllFields` which **adds** the **'Begin, Finish & Notes'** fields to the csv export. Default: `false`  
For now this only does something when you specified `--format "csv"` but this will change in the future and apply to all exports.



### Issues 
If you find issues or bugs, by all means please open an issue with a description and I will take a look at it as soon as I can. 

---
---
# ORIGINAL README FROM THE FORKED PROJECT

---

ZEIT
----

![zeit](documentation/zeit.png)

Zeit, erfassen. A command line tool for tracking time spent on tasks & projects.

[Get some more info on why I build this
here](https://マリウス.com/zeit-erfassen-a-cli-activity-time-tracker/).

[Download the latest version for macOS, Linux, FreeBSD, NetBSD, OpenBSD & Plan9
here](https://github.com/mrusme/zeit/releases/latest).


## Build

```sh
make
```

**Info**: This will build using the version 0.0.0. You can prefix the `make` 
command with `VERSION=x.y.z` and set `x`, `y` and `z` accordingly if you want 
the version in `zeit --help` to be a different one.


## Usage

![zeit](documentation/header.jpg)

Please make sure to `export ZEIT_DB=~/.config/zeit.db` (or whatever location 
you would like to have the zeit database at).

*zeit*'s data structure contains of the following key entities: `project`, 
`task` and `entry`. An `entry` consists of a `project` and a `task`. These
don't have to pre-exist and can be created on-the-fly inside a new `entry` using
e.g. `zeit track --project "New Project" --task "New Task"`. In order to
configure them, the `zeit project` and the `zeit task` commands can be utilised.


### Projects

A project can be configured using `zeit project`:

```sh
zeit project --help
```

#### Examples:

Set the project color to a hex color code, allowing `zeit stats` to display
information in that color (if your terminal supports colours):

```sh
zeit project --color '#d3d3d3' "cool project"
```


### Task

A task can be configured using `zeit task`:

```sh
zeit task --help
```

#### Examples:

Setting up a Git repository to have commit messages automatically imported
into the activity notes when an activity is finished:

```sh
zeit task --git ~/my/git/repository "development"
```

**Info:** You will have to have the `git` binary available in your `PATH` for 
this to work. *zeit* automatically limits the commit log to the exact time of 
the activity's beginning- and finish-time. Commit messages before or after these 
times won't be imported.


### Track activity

```sh
zeit track --help
```

#### Examples:

Begin tracking a new activity and reset the start time to 15 minutes ago:

```sh
zeit track --project project --task task --begin -0:15
```


### Show current activity

```sh
zeit tracking
```


### Finish tracking activity

```sh
zeit finish --help
```

#### Examples:

Finish tracking the currently tracked activity without adding any further info:

```sh
zeit finish
```

Finish tracking the currently tracked activity and change its task:

```sh
zeit finish --task other-task
```

Finish tracking the currently tracked activity and adjust its start time to 
4 PM:

```sh
zeit finish --begin 16:00
```


### List tracked activity

```sh
zeit list --help
```

#### Examples:

List all tracked activities:

```sh
zeit list
```

List all tracked activities since a specific date/time:

```sh
zeit list --since "2020-10-14T00:00:01+01:00"
```

List all tracked activities and add the total hours:

```sh
zeit list --total
```

List only projects and tasks (relational):

```sh
zeit list --only-projects-and-tasks
```

List only projects and tasks (relational) that were tracked since a specific 
date/time:

```sh
zeit list --only-projects-and-tasks --since "2020-10-14T00:00:01+01:00"
```


### Display/update activity

```sh
zeit entry --help
```

#### Examples:

Display a tracked activity:

```sh
zeit entry 14037730-5c2d-44ff-b70e-81f1dcd4eb5f
```

Update a tracked activity:

```sh
zeit entry --finish "2020-09-02T18:16:00+01:00" 14037730-5c2d-44ff-b70e-81f1dcd4eb5f
```


### Erase tracked activity

```sh
zeit erase --help
```

#### Examples:

Erase a tracked activity by its internal ID:

```sh
zeit erase 14037730-5c2d-44ff-b70e-81f1dcd4eb5f
```


### Statistics

![zeit stats](documentation/zeit_stats.jpg)

```sh
zeit stats
```


### Import tracked activities

```sh
zeit import --help
```

The following formats are supported as of right now:

#### `tyme`: Tyme 3 JSON

It is possible to import JSON exports from [Tyme 3](https://www.tyme-app.com). 
It is important that the JSON is exported with the following options set/unset:

![Tyme 3 JSON export](documentation/tyme3json.png)

- `Start`/`End` can be set as required
- `Format` has to be `JSON`
- `Export only unbilled entries` can be set as required
- `Mark exported entries as billed` can be set as required
- `Include non-billable tasks` can be set as required
- `Filter Projects & Tasks` can be set as required
- `Combine times by day & task` **must** be unchecked

During import, *zeit* will create SHA1 sums for every Tyme 3 entry, which 
allows it to identify every imported activity. This way *zeit* won't import the 
exact same entry twice. Keep this in mind if you change entries in Tyme and 
then import them again into *zeit*.

#### Examples:

Import a Tyme 3 JSON export:

```sh
zeit import --format tyme ./tyme.export.json
```


### Export tracked activities

```sh
zeit export --help
```

The following formats are supported as of right now:

#### `zeit`: *zeit* JSON

The *zeit* internal JSON format. Basically a dump of the database including
only tracked activities.

#### `tyme`: Tyme 3 JSON

It is possible to export JSON compatible to the Tyme 3 JSON format. Fields that
are not available in *zeit* will be filled with dummy values, e.g.
`Billing: "UNBILLED"`.

#### Examples:

Export a Tyme 3 JSON:

```sh
zeit export --format tyme --project "my project" --since "2020-04-01T15:04:05+07:00" --until "2020-04-04T15:04:05+07:00"
```

## Integrations

Here are a few integrations and extensions built by myself as well as other 
people that make use of `zeit`:

- [`zeit-waybar-bemenu.sh`](https://github.com/mrusme/zeit/blob/main/extras/zeit-waybar-bemenu.sh), 
  a script for integrating `zeit` into
  [waybar](https://github.com/Alexays/Waybar), using
  [bemenu](https://github.com/Cloudef/bemenu)
- [`zeit-waybar-wofi.sh`](https://github.com/mrusme/zeit/blob/main/extras/zeit-waybar-wofi.sh), 
  a script for integrating `zeit` into
  [waybar](https://github.com/Alexays/Waybar), using
  [wofi](https://hg.sr.ht/~scoopta/wofi)
- [`zeit.1m.sh`](https://github.com/mrusme/zeit/blob/main/extras/zeit.1m.sh), 
  an [`xbar`](https://github.com/matryer/xbar) plugin for `zeit`
- [`zeit-status.sh`](https://github.com/khughitt/dotfiles/blob/master/polybar/scripts/zeit-status.sh), 
  a [Polybar](https://github.com/polybar/polybar) integration for `zeit` by 
  [@khughitt](https://github.com/khughitt) 
  (see [#1](https://github.com/mrusme/zeit/issues/1))
- your link here, feel free to PR! :-)
