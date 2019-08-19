package rotatedlog

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

/*
Logger - log file object
*/
type Logger struct {
	// то имя, которое задаем при инициализации
	FileName string
	// то имя, которое получаем генерацией
	filename string
	// через сколько часов обновлять имя файла
	rotateAfter float64

	file *os.File
	mu   sync.Mutex

	activeDate time.Time
}

const (
	fileNameTemplate = "2006-01-02-15"
)

// Init - Initialize new log object
func Init(fname string, hours int) (*Logger, error) {
	res := &Logger{
		rotateAfter: float64(hours),
	}
	res.createFileName(fname)
	err := res.openActiveFile()
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Close log file
func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	// если файл был открыт - закроем
	if l.file != nil {
		l.file.Close()
	}
}

// создаем настоящее имя файла (включая дату и время)
func (l *Logger) createFileName(fname string) {
	// получим текущую дату и разложим на составляющие
	l.FileName = fname
	l.activeDate = time.Now()

	// создадим имя активного файла
	dn, fn := filepath.Split(fname)
	var ext = filepath.Ext(fn)
	var name = fn[0:(len(fn) - len(ext))]
	name = fmt.Sprintf("%s_%s%s", name, l.activeDate.Format(fileNameTemplate), ext)

	l.filename = filepath.Join(dn, name)
}

// Write запись в файл строки (форматируем в "type YYYY-MM-DD HH:MM:SS <строка>")
func (l *Logger) Write(logtype string, s string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// тут проверим нужно ли поменять файл
	if l.rotateNeeded() {
		l.rotate()
	}
	var err error
	// создадим файл и откроем его если файл не был создан
	if l.file == nil {
		// создать новый файл
		l.file, err = os.OpenFile(l.filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			return err
		}
	}
	// собственно запись
	str := fmt.Sprintf("%s %s %s\n", logtype, time.Now().Format("2006-01-02 15:04:05"), s)
	_, err = l.file.WriteString(str)

	return err
}

// поменять имя файла
func (l *Logger) rotate() error {
	// закроем файл (если он был открыт)
	if l.file != nil {
		l.file.Close()
	}
	// теперь соберем новое имя
	dn, fn := filepath.Split(l.FileName)
	var ext = filepath.Ext(fn)
	var name = fn[0:(len(fn) - len(ext))]
	name = fmt.Sprintf("%s_%s%s", name, l.activeDate.Format(fileNameTemplate), ext)

	l.filename = filepath.Join(dn, name)
	// и откроем его
	_, err := os.Stat(l.filename)
	if os.IsNotExist(err) {
		// создать новый файл
		l.file, err = os.OpenFile(l.filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		return err
	}
	// открыть существующий файл
	l.file, err = os.OpenFile(l.filename, os.O_APPEND|os.O_WRONLY, 0666)
	return err
}

// нужно или нет менять файл
func (l *Logger) rotateNeeded() bool {
	t := time.Now()
	// year, month, day := t.Date()
	if t.Sub(l.activeDate).Hours() >= l.rotateAfter {
		// if (year != l.year) || (month != l.month) || (day != l.day) {
		// сохраним дату для создания файла
		l.activeDate = t
		return true
	}
	return false
}

// откроем файл или создадим его если его нет
func (l *Logger) openActiveFile() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, err := os.Stat(l.filename)
	if os.IsNotExist(err) {
		// если файла нет - то не создаем его
		l.file = nil
		return nil
	}
	// открыть существующий файл
	l.file, err = os.OpenFile(l.filename, os.O_APPEND|os.O_WRONLY, 0666)
	return err
}
