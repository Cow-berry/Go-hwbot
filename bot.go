package main

import (
    "github.com/Syfaro/telegram-bot-api"
    "io/ioutil"
    "log"
    "fmt"
    "os"
    str "strings"
    "strconv"
    "config"
    "time"
    "telegram_calender"
    "net/http"
    "io"
    // "database/sql"
    // _ "github.com/mattn/go-sqlite3"c
)

// TODO аватарка, разные языки

var bot *tgbotapi.BotAPI

func is_admin(id int64) bool {
    if id == config.Admin {return true}
    return false
}

func in_slice(a string, s []string) bool {
    for _, b := range s {
        if b == a {return true}
    }
    return false
}

func get_date(add ...int) string {
    days := 0
    if add != nil {days = add[0]}
    year, month, day := time.Now().AddDate(0, 0, days).Date()
    return fmt.Sprintf("%v.%02d.%02d", year, month, day)
}

func update_timetable() {
    table, _ := read("timetable.txt")
    table_strings := str.Split(table, "\n")
    timetable = [][]string{}
    for _, table_string := range table_strings {
        if table_string == "" {continue}
        timetable = append(timetable, str.Split(table_string, " "))
    }
}

var timetable [][]string

var users = map[int64]string{int64(310802215) : "main_menu"}
//reads txt files and returns all the strings combined in one
func Log(id int64, text string, username string) {
    write("log.txt", fmt.Sprintf("%v(%v): %v\n", username, id, text), false)
}

func write(name string, text string, rew bool) error {
    file, err := os.OpenFile(name, os.O_APPEND|os.O_WRONLY, 0600)
    if rew {
        file, err = os.Create(name)
    }
    defer file.Close()
    if (err != nil) {return err}
    _, err = file.Write([]byte(text))
    return err
}

func read(name string) (string, error) {
    data, err := ioutil.ReadFile(name)
    if err != nil {
        return "oops, no data found", err
    }
    return string(data), nil
}

func keyboard(id int64, text string, rows [][]string) {
    msg := tgbotapi.NewMessage(id, text)
    keyboard := tgbotapi.NewReplyKeyboard()
    for _, row := range [][]string(rows) {
        krow := tgbotapi.NewKeyboardButtonRow()
        for _, button := range []string(row) {
            krow = append(krow, tgbotapi.NewKeyboardButton(string(button)))
        }
        keyboard.Keyboard = append(keyboard.Keyboard, krow)
    }
    msg.ReplyMarkup = keyboard
    bot.Send(msg)
}

var teachers = map[string]string{"Алгебра":"", "Английский":"", "Геометрия":"", "История":"", "Литература":"", "Мат.Ан.":"", "МатематикаЕГЭ":"", "Обществознание":"", "Программирование":"", "Русский":"", "Статистика":"", "Физ-ра":"", "Физика":""}

var subj_list = []string{"Алгебра", "Английский", "Геометрия", "История", "Литература", "Мат.Ан.", "МатематикаЕГЭ", "Обществознание", "Программирование", "Русский", "Статистика", "Физ-ра"}

var subj = [][]string{{"Алгебра", "Английский", "Геометрия"}, {"История", "Литература", "Мат.Ан."}, {"МатематикаЕГЭ", "Обществознание", "Программирование"}, {"Русский", "Статистика", "Физ-ра"}, {"Физика", "старт"}}


var konspekt_subj = [][]string{{"Мат.Ан.", "Геометрия"}, {"Алгебра", "Обществознание"}, {"старт"}} // заполнить

var colloq = map[string]string{"ege": "МатематикаЕГЭ", "en" : "Английский", "geo" : "Геометрия", "his": "История", "lit": "Литература", "ma": "Мат.Ан.", "pe": "Физ-ра", "ph": "Физика", "pr": "Программирование", "ru": "Русский", "ss": "Обществознание", "st": "Статистика"}

// const main_menu_keys = [][][]string{{{"дз", "расписание", "старт", "помощь"}, {"конспект", "учителя", "пожелание", "настройки"}},
                                    // {{"get hw", "get tt", "start", "help"}, {"konspekt", "teachers", "wish", "settings"}},
                                // }

// const hw_menu_keys = [][][]string{{{"завтра", "сегодня", "всё", "по предмету"}},
                                    // {"next", "present", "all", "course's"},
                                // }


func main_menu(id int64) {
    keyboard(id, "Главное меню", [][]string{{"дз", "старт",  "расписание"}, {"дежурные", "конспект", "учителя"}, {"пожелание"}})
}


func hw_menu(id int64) {
    keyboard(id, "На какой временной период:", [][]string{{"сегодня", "завтра", "всё"}, {"по предмету", "старт", "на дату"}})
}

func sub_menu(id int64) {
    keyboard(id, "Выберите предмет:", subj)
}

func konspekt_menu(id int64) {
    keyboard(id, "Выберите предмет:", konspekt_subj)
}

func start(id int64) {
    users[id] = "main_menu"
    main_menu(id)
}

func convert(s string) string {
    if in_slice(s, []string{"дз", "hw", "kotitehtävät", "hausaufgaben"}) {return "дз"}
    if in_slice(s, []string{"расписание", "timetable", "aikataulu", "zeitplan"}) {return "расписание"}
    if in_slice(s, []string{"пожелание", "wish", "toive", "wunsch"}) {return "пожелание"}
    if in_slice(s, []string{"всё", "all", "kaikki", "alles"}) {return "всё"}
    if in_slice(s, []string{"завтра", "tomorrow", "huomenna", "morgen"}) {return "завтра"}
    if in_slice(s, []string{"сегодня", "today", "tänään", "heute"}) {return "сегодня"}
    return s
}

func reply(update tgbotapi.Update) {
    id := update.Message.Chat.ID
    // log.Print(timetable)
    doc := ""
    if update.Message.Photo != nil {
        doc = "photo"
    } else if update.Message.Document != nil {
        doc = "doc"
    }

    if doc != "" {
        // log.Print("doc")
        var resp tgbotapi.File
        var path, name, format string
        var buf []byte
        if doc == "photo" {
            photo := *update.Message.Photo
            caption := str.Split(update.Message.Caption, " ")
            name = caption[0]
            format = caption[1]
            resp, _ = bot.GetFile(tgbotapi.FileConfig{photo[len(photo)-1].FileID})
            path = "hw/files/"
            buf = make([]byte, photo[len(photo)-1].FileSize)
        } else {
            var ok bool
            name, ok = colloq[update.Message.Caption]
            if !ok {return}
            document := *update.Message.Document
            format = "pdf"
            resp, _ = bot.GetFile(tgbotapi.FileConfig{document.FileID})
            path = "lecture notes/"
            buf = make([]byte, document.FileSize)
        }
        r, _ := http.Get("https://api.telegram.org/file/bot"+config.Token+"/"+resp.FilePath)
        io.ReadFull(r.Body, buf)
        file, _ := os.Create(path+name+"."+format)
        file.Write(buf)
        file.Close()
        start(id)
        return
    }
    //
    // if update.Message.Photo != nil ||  {
    //     log.Print("photo")
    //     log.Print(update.Message.Caption)
    //     if !( is_admin(id)) {
    //         log.Print("not load")
    //         return
    //     }
    //     var format, name string
    //     file, _ := os.Create("hw/files/"+name+"."+format)
    // }
    // if update.Message.Document != nil {
    //     r, _ := http.Get("https://api.telegram.org/file/bot"+config.Token+"/"+resp.FilePath)
    //     io.ReadFull(r.Body, buf)
    //     file, _ := os.Create("lecture notes/"+sub+".pdf")
    //     file.Write(buf)
    //     file.Close()
    //     start(id)
    //     return
    // }

    // log.Print(timetable)
    text := update.Message.Text
    if text == "test" || text == "тест" {
        msg := tgbotapi.NewMessage(id, "test")
        msg.ReplyMarkup = telegram_calender.Calender(2, 2020)
        bot.Send(msg)
        // log.Print(telegram_calender.Month_calender(2, 2020))
        return
    }

    logs <- log_msg{id, text, update.Message.Chat.UserName}
    // msg := tgbotapi.NewMessage(int64(310802215), "starting well")
    // bot.Send(msg)
    if text == "старт" || text == "/start" {
        start(id)
        return
    }

    if is_admin(id) {
        if stext := str.Split(text, " "); stext[0] == "write" {
            write_hw(id, text) // TODO
            return
        } else if stext[0] == "teach" {
            if len(stext) < 3 {return}
            if _, ok := teachers[colloq[stext[1]]]; ok {
                teachers[colloq[stext[1]]] = str.Join(stext[2:], " ")
            }
            return
        }
        stext := str.Split(str.Split(text, "\n")[0], " ")
        if stext[0] == "table" {
            change_timetable(id, text)
            start(id)
            return
        } else if stext[0] == "log" {
            bot.Send(tgbotapi.NewDocumentUpload(id, "bot.log"))
            start(id)
            return
        }
    }
    _, ok := users[id]
    if !ok {return}
    // log.Print(text, id)
    text = convert(text)
    switch text {
    case "дз":
        hw_menu(id)
        users[id] = "hw_menu"
    case "расписание":
        tt(id)
        start(id)
    case "пожелание":
        keyboard(id, "Наберите пожелание:", [][]string{{"Cancel"}})
        users[id] = "wish"
    case "всё":
        for i:=0;i<7;i++ {
            hw(id, get_date(i), "")
        }
        start(id)
    case "завтра":
        hw(id, "tomorrow", "")
        start(id)
    case "сегодня":
        hw(id, "today", "")
        start(id)
    case "по предмету":
        users[id] = "sub_menu"
        sub_menu(id)
    case "конспект":
        users[id] = "konspekt"
        konspekt_menu(id)
    case "на дату":
        msg := tgbotapi.NewMessage(id, "Выберите дату:")
        year, month, _ := time.Now().Date()
        msg.ReplyMarkup = telegram_calender.Calender(int(month), year)
        new_msg, _ := bot.Send(msg)
        prev_calender, ok := calenders[id]
        if ok {
            bot.DeleteMessage(tgbotapi.DeleteMessageConfig{id, prev_calender.msg_id})
        }
        calenders[id] = hw_calender{new_msg.MessageID, int(month), year}
        start(id)
    // case "опросы":
    //     bot.Send(tgbotapi.NewMessage(id, "Пока опросов нет"))
    //     start(id)
    case "учителя":
        text = ""
        for sub, teach := range teachers {
            text += fmt.Sprintf("⌈%s⌋ — %s\n", sub, teach)
        }
        bot.Send(tgbotapi.NewMessage(id, text))
    case "дежурные":
        duty, _ := read("duty.txt")
        bot.Send(tgbotapi.NewMessage(id, duty))
    default:
        switch users[id]{
        case "sub_menu":
            hw(id, "", text)
        case "date":
            hw(id, text, "")
        case "wish":
            if text != "Cancel" {
                wishes_chan <- wishes{update.Message.MessageID, id, update.Message.From.FirstName, text}
            }
        case "konspekt":
            bot.Send(tgbotapi.NewDocumentUpload(id, "lecture notes/"+text+".pdf"))
        default:
            bot.Send(tgbotapi.NewMessage(id, "Неизвестная команда"))
        }
         start(id)
    }
}

func sum (array []string) string {
    res := ""
    for _, elem := range array {
        res += elem
    }
    return res
}

func build_timtable(timetable []string) string {
    var row []string
    ntimetable := ""
    for _, day := range timetable {
        row = []string{}
        for _, subject := range str.Split(day, " ") {
            row = append(row, colloq[subject])
        }
        ntimetable += str.Join(row, " ") + "\n"
    }
    return ntimetable
}

func change_timetable(id int64, text string) {
    stext := str.Split(text, "\n")
    defining := str.Split(stext[0], " ")
    // var row []string
    if len(defining) == 1 {
        write("timetable.txt", build_timtable(stext[1:]), true)
    } else if len(defining) == 2 {
        day, _ := strconv.Atoi(defining[1])
        ptimetable, _ := read("timetable.txt")
        sptimetable := str.Split(ptimetable, "\n")
        row := []string{}
        for _, sub := range str.Split(stext[1], " ") {
            row = append(row, colloq[sub])
        }
        sptimetable[day-1] = str.Join(row, " ")
        write("timetable.txt", str.Join(sptimetable, "\n"), true)
    } else if len(defining) == 4 {
        day, _ := strconv.Atoi(defining[1])
        pair, _ := strconv.Atoi(defining[2])
        ptimetable, _ := read("timetable.txt")
        sptimetable := str.Split(ptimetable, "\n")
        // log.Print(sptimetable)
        ssptimetable := str.Split(sptimetable[day-1], " ")
        ssptimetable[pair-1] = colloq[defining[3]]
        sptimetable[day-1] = str.Join(ssptimetable, " ")
        write("timetable.txt", str.Join(sptimetable, "\n"), true)
    } else {
        bot.Send(tgbotapi.NewMessage(id, "Проверьте запись изменения"))
        return
    }
    update_timetable()
}

type hw_calender struct {
    msg_id int
    month int
    year int
}

var calenders = map[int64]hw_calender{}

func answer(update tgbotapi.Update) {
    data := update.CallbackQuery.Data
    id := int64(update.CallbackQuery.From.ID)
    switch data{
    case "ignore":
        return
    case "next":
        change_calender(id, 1)
    case "prev":
        change_calender(id, -1)
    default:
        data, _ := strconv.Atoi(data)
        calender := calenders[id]
        // log.Print(fmt.Sprintf("%v %v.%02d.%02d", id, calender.year, calender.month,data))
        hw(id, fmt.Sprintf("%v.%02d.%02d", calender.year, calender.month, data), "")
    }
    bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
}

func change_calender(id int64, shift int) {
    calender := calenders[id]
    month := calender.month + shift
    year := calender.year
    if month == 0 {
        month = 12
        year--
    } else if month == 13 {
        month = 1
        year++
    }
    // log.Print("editing")
    // log.Print(tgbotapi.NewEditMessageReplyMarkup(id, calender.msg_id, telegram_calender.Calender(month, year)))
    bot.Send(tgbotapi.NewEditMessageReplyMarkup(id, calender.msg_id, telegram_calender.Calender(month, year)))
    calenders[id] = hw_calender{calender.msg_id, month, year}

}

func get_next_date(sub string) string {
    wd := time.Now().Weekday()
    wd++
    var i int = 1
    if wd == time.Sunday {
        wd = time.Monday
    }
    for ;i<=6;i++ {
        // log.Print(timetable[(int(wd)-1)%7], in_slice(sub, timetable[(int(wd)-1)%7-1]))
        // log.Print(sub, timetable[(int(wd)-1)%6], (int(wd)-1)%6)
        if in_slice(sub, timetable[(int(wd)-1)%6]) {return get_date(i)}
        wd++
        // i++
        if wd == time.Sunday {
            wd = time.Monday
            i--
        }
    }
    return "не найдено предмета " + sub
}

func write_hw(id int64, text string) {
    var msg_text string
    stext := str.Split(text, "\n")
    cmd := str.Split(stext[0], " ")
    date := cmd[1]
    sub := cmd[2]
    if date == "next" {date = get_next_date(colloq[sub])}
    if _, ok := colloq[sub]; date[:0] == "н" || !ok {
        msg_text = "не найдено предмета " + sub
    } else {
        write("hw/"+colloq[sub]+"/"+date+".txt", text[len(stext[0])+1:], true) // +1 to avoid writing the first '\n'
        msg_text = "done"
    }
    bot.Send(tgbotapi.NewMessage(id, msg_text))
}

func hw(id int64, date string, sub string) {
    if sub != "" {
        files, _ := ioutil.ReadDir(fmt.Sprintf("hw/%v", sub))
        for _,  f := range files {
            hw_text, _ := read("hw/" + sub + "/" + f.Name())
            splitted_hw_text := str.Split(hw_text, "\n")
            final_hw_text := ""
            var file []string
            for _, hw_string := range splitted_hw_text {
                if in_slice("file", str.Split(hw_string, " ")) {
                    file = str.Split(str.Split(hw_string, " ")[1], ".")
                    if file[1] == "png" || file[1] == "jpg" {
                        bot.Send(tgbotapi.NewPhotoUpload(id, "hw/files/"+file[0]+"."+file[1]))
                    } else {
                        bot.Send(tgbotapi.NewDocumentUpload(id, "hw/files/"+file[0]+"."+file[1]))
                    }
                } else {
                    final_hw_text += "` ⠀`" + hw_string + "\n"
                }
            }

            msg := tgbotapi.NewMessage(id, f.Name()[:len(f.Name())-5] + "\n" + final_hw_text)
            msg.ParseMode = "Markdown"
            if msg.Text != "" {
                bot.Send(msg)
            }
        }
        return
    }
    if date == "tomorrow" {
        date = get_date(1)
    } else if date == "today" {
        date = get_date()
    }
    dl := str.Split(date, ".") //date_list
    y, _ := strconv.ParseInt(dl[0], 10, 0)
    m, _ := strconv.ParseInt(dl[1], 10, 0)
    d, _ := strconv.ParseInt(dl[2], 10, 0)
    date_obj := time.Date(int(y), time.Month(m), int(d), 0, 0, 0, 0, time.UTC)
    var subs []string
    if date_obj.Before(time.Now()) {
        subs = subj_list
    } else {
        weekday := (d + m + y + y/4 + 21 - 2) % 7
        if weekday == 6 {weekday=0} // Sunday -- 0, timetable[0] -- Monday  +-1 -> still Monday ; Saturday -- 6, timetable[6-1] -- Saturday
        subs = timetable[weekday]
    }
    var hw_texts string = ""
    // log.Print(subs, weekday)
    for _, sub := range subs {
        hw_text, err := read("hw/"+sub+"/"+date+".txt")
        final_hw_text := ""
        var file []string
        splitted_hw_text := str.Split(hw_text, "\n")
        for _, hw_string := range splitted_hw_text {
            if in_slice("file", str.Split(hw_string, " ")) {
                file = str.Split(str.Split(hw_string, " ")[1], ".")
                if file[1] == "png" || file[1] == "jpg" {
                    bot.Send(tgbotapi.NewPhotoUpload(id, "hw/files/"+file[0]+"."+file[1]))
                } else {
                    bot.Send(tgbotapi.NewDocumentUpload(id, "hw/files/"+file[0]+"."+file[1]))
                }
            } else {
                final_hw_text += "` ⠀`" + hw_string + "\n"
            }
        }
        if err == nil {
            hw_texts += sub + "\n" + final_hw_text
        }
    }
    msg := tgbotapi.NewMessage(id, date + "\n" + hw_texts)
    msg.ParseMode = "Markdown"
    if msg.Text != "" {
        bot.Send(msg)
    }

    // hw_text, err := get_hw(date, sub)
    // msg := tgbotapi.NewMessage(id, hw_text)
    // if (err != nil) {
    //     msg.Text = "Не удалось получить дз по предмету " + sub + " на "  + get_date()
    //     log.Print(err)
    // }
    // bot.Send(msg)
    // start(id)
}

var days_of_week_rus = []string{"Понедельник", "Вторник", "Среда", "Четверг", "Пятница", "Суббота"}

func tt(id int64) {
    // tt, _ := read("timetable.txt")
    var tt string = ""
    // var env string = "```"
    wd := 0
    awd := time.Now().Weekday()//actual weekday
    if awd == time.Sunday {awd++}
    for n, line := range timetable {
        tt += days_of_week_rus[wd] + ":\n"
        wd++
        for _, sub := range line {
            if n == int(awd - 1) {
                tt += "`   ⠀`" + sub + "\n" // there is a special "empty" character at the end of the gap. It's not a space, so text fairly make 4 equal-sized spaces
            } else {tt += "`    " + sub + "`\n"}
        }
        }
    msg := tgbotapi.NewMessage(id, tt)
    msg.ParseMode = "Markdown"
    bot.Send(msg)
}

func wish(id int64) {
    bot.Send(tgbotapi.NewMessage(id, "wish test"))
}

type log_msg struct {
    id int64
    text string
    username string
}
type wishes struct {
    id int
    uid int64
    name string
    text string
}
var logs chan log_msg
var wishes_chan chan wishes
var err error
func main() {


    //initialize bot with token
    // bot, err := tgbotapi.NewBotAPI(config.Token)
    // if err != nil {
    //     panic(err)
    // }
    //Устанавливаем время обновления
    // log.Print(timetable)
    update_timetable()
    bot, err = tgbotapi.NewBotAPI(config.Token)
    // bot.Debug = true
    if (err != nil) {panic(err)}
    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60
    updates, _ := bot.GetUpdatesChan(u)
    bot.Send(tgbotapi.NewMessage(int64(310802215), "started well"))
    //create a gourutine for every message
    // log.Print("started")
    logs = make(chan log_msg, 100)
    go func() {
        logf, _ := os.OpenFile("bot.log", os.O_APPEND|os.O_WRONLY, 0644)
        log.SetOutput(logf)
        defer logf.Close()
        for single_log := range logs {
            // log.Print("it logs")
            log.Print(fmt.Sprintf("%s(%d) %s", single_log.username, single_log.id, single_log.text))
            // Log()
        }
        // log.Print("it stopped")
    }()


    wish_bot, _ := tgbotapi.NewBotAPI(config.Wish_token)
    // wish_bot.Debug = true
    wish_updates, _ := wish_bot.GetUpdatesChan(u)
    wishes_chan = make(chan wishes)
    go func() {
        for wish := range wishes_chan {
            wish_bot.Send(tgbotapi.NewMessage(config.Admin, strconv.FormatInt(int64(wish.id), 10) + " " + strconv.FormatInt(wish.uid, 10) + " " + wish.name + "\n" + wish.text))
        }
    }()

    go func() {
        for update := range wish_updates {
            if update.Message.Chat.ID != config.Admin || update.Message.ReplyToMessage == nil {continue}
            info := str.Split(update.Message.ReplyToMessage.Text, " ")[:2]
            id, _ := strconv.ParseInt(info[0], 10, 0)
            uid, _ := strconv.ParseInt(info[1], 10, 64)
            msg := tgbotapi.NewMessage(uid, update.Message.Text)
            msg.ReplyToMessageID = int(id)
            bot.Send(msg)
        }
    }()

    for update := range updates {
        if update.CallbackQuery != nil {
            // log.Print("callback")
            go answer(update)
        } else {
            go reply(update)
        }
    }

}
