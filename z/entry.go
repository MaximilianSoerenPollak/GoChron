package z

import (
  "errors"
  "strings"
  "time"
  "fmt"
  "github.com/gookit/color"
)

type Entry struct {
  ID      string      `json:"-"`
  Begin   time.Time   `json:"begin,omitempty"`
  Finish  time.Time   `json:"finish,omitempty"`
  Project string      `json:"project,omitempty"`
  Task    string      `json:"task,omitempty"`
  User    string      `json:"user,omitempty"`
}

func NewEntry(
  id string,
  begin string,
  finish string,
  project string,
  task string,
  user string) (Entry, error) {
  var err error

  newEntry := Entry{}

  newEntry.ID = id
  newEntry.Project = project
  newEntry.Task = task
  newEntry.User = user

  _, err = newEntry.SetBeginFromString(begin)
  if err != nil {
    return Entry{}, err
  }

  _, err = newEntry.SetFinishFromString(finish)
  if err != nil {
    return Entry{}, err
  }

  if newEntry.IsFinishedAfterBegan() == false {
    return Entry{}, errors.New("beginning time of tracking cannot be after finish time")
  }

  return newEntry, nil
}

func (entry *Entry) SetIDFromDatabaseKey(key string) (error) {
  splitKey := strings.Split(key, ":")

  if len(splitKey) < 3 || len(splitKey) > 3 {
    return errors.New("not a valid database key")
  }

  entry.ID = splitKey[2]
  return nil
}

func (entry *Entry) SetBeginFromString(begin string) (time.Time, error) {
  var beginTime time.Time
  var err error

  if begin == "" {
    beginTime = time.Now()
  } else {
    beginTime, err = ParseTime(begin)
    if err != nil {
      return beginTime, err
    }
  }

  entry.Begin = beginTime
  return beginTime, nil
}

func (entry *Entry) SetFinishFromString(finish string) (time.Time, error) {
  var finishTime time.Time
  var err error

  if finish != "" {
    finishTime, err = ParseTime(finish)
    if err != nil {
      return finishTime, err
    }
  }

  entry.Finish = finishTime
  return finishTime, nil
}

func (entry *Entry) IsFinishedAfterBegan() (bool) {
  return (entry.Finish.IsZero() || entry.Begin.Before(entry.Finish))
}

func (entry *Entry) GetOutputForTrack(isRunning bool, wasRunning bool) (string) {
  var outputPrefix string = ""
  var outputSuffix string = ""

  now := time.Now()
  trackDiffNow := now.Sub(entry.Begin)
  trackDiffNowOut := time.Time{}.Add(trackDiffNow)

  if isRunning == true && wasRunning == false {
    outputPrefix = "began tracking"
  } else if isRunning == true && wasRunning == true {
    outputPrefix = "tracking"
    outputSuffix = fmt.Sprintf(" for %sh", color.FgLightWhite.Render(trackDiffNowOut.Format("15:04")))
  } else if isRunning == false && wasRunning == false {
    outputPrefix = "tracked"
  }

  if entry.Task != "" && entry.Project != "" {
    return fmt.Sprintf("%s %s %s on %s%s\n", CharTrack, outputPrefix, color.FgLightWhite.Render(entry.Task), color.FgLightWhite.Render(entry.Project), outputSuffix)
  } else if entry.Task != "" && entry.Project == "" {
    return fmt.Sprintf("%s %s %s%s\n", CharTrack, outputPrefix, color.FgLightWhite.Render(entry.Task), outputSuffix)
  } else if entry.Task == "" && entry.Project != "" {
    return fmt.Sprintf("%s %s task on %s%s\n", CharTrack, outputPrefix, color.FgLightWhite.Render(entry.Project), outputSuffix)
  }

  return fmt.Sprintf("%s %s task%s\n", CharTrack, outputPrefix, outputSuffix)
}

func (entry *Entry) GetOutputForFinish() (string) {
  var outputSuffix string = ""

  trackDiff := entry.Finish.Sub(entry.Begin)
  trackDiffOut := time.Time{}.Add(trackDiff)

  outputSuffix = fmt.Sprintf(" for %sh", color.FgLightWhite.Render(trackDiffOut.Format("15:04")))

  if entry.Task != "" && entry.Project != "" {
    return fmt.Sprintf("%s finished tracking %s on %s%s\n", CharFinish, color.FgLightWhite.Render(entry.Task), color.FgLightWhite.Render(entry.Project), outputSuffix)
  } else if entry.Task != "" && entry.Project == "" {
    return fmt.Sprintf("%s finished tracking %s%s\n", CharFinish, color.FgLightWhite.Render(entry.Task), outputSuffix)
  } else if entry.Task == "" && entry.Project != "" {
    return fmt.Sprintf("%s finished tracking task on %s%s\n", CharFinish, color.FgLightWhite.Render(entry.Project), outputSuffix)
  }

  return fmt.Sprintf("%s finished tracking task%s\n", CharFinish, outputSuffix)
}
