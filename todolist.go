package main

import (
	"database/sql"  //database/sql paketi, SQL veritabanı işlemleri için temel işlevselliği sağlar
	"fmt"           //paketi, çeşitli formatlama işlemleri ve çıktı yöntemleri sağlar.
	"html/template" // paketi, HTML şablonları oluşturmak için kullanılır
	"net/http"      //paketi, HTTP sunucu ve istemci işlemleri için temel işlevselliği sağlar.
	"strconv"       //paketi, string ve sayı arasında dönüşümler yapmak için kullanılır.

	_ "github.com/lib/pq" // PostgreSQL veritabanı sürücüsünü yükler.
)

// Task struct'u, her bir görevin temsil edilmesi için kullanılır.
type Task struct {
	ID          int
	Title       string
	Description string
	Archived    sql.NullBool // Arşiv durumu
}

// Page struct'u, HTML sayfasının başlığını ve görevlerini içerir.
type Page struct {
	PageTitle string
	Tasks     []Task // Bu yapı, daha sonra HTML şablonunda kullanılarak görevleri listeleyen bir web sayfası oluşturulur.
}

var db *sql.DB // PostgreSQL veritabanı bağlantısı

func init() {
	// PostgreSQL veritabanına bağlan
	connStr := "user=postgres password=merve dbname=go-todo sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Veritabanına bağlanma hatası:", err)
		return
	}

	// Veritabanı bağlantısını kontrol et
	err = db.Ping()
	if err != nil {
		fmt.Println("Veritabanı bağlantısı başarısız:", err)
		return
	}

	fmt.Println("PostgreSQL veritabanına başarıyla bağlandı")
}

// indexHandler, ana sayfayı gösterir ve mevcut görevleri listeler.
func indexHandler(w http.ResponseWriter, r *http.Request) { //, HTTP yanıtının yazılacağı yerdir>>Responsewirter HTTP isteğini temsil eder.>>Request
	// Veritabanından görevleri al
	tasks, err := getTasks()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// index.html dosyasını kullanarak bir template oluştur
	tmpl, err := template.New("index").Parse(`
    <!DOCTYPE html>
    <html lang="en">
    
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>{{.PageTitle}}</title>
        <!-- Bootstrap CSS -->
        <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css">
        <style>
            /* Custom Styles */
            body {
                background-color: #ffc0cb; /* Pembe arka fon */
                color: #2c3e50;
                font-family: 'Arial', sans-serif;
                margin: 20px;
            }
    
            .container {
                margin-top: 50px;
            }
    
            h1 {
                color: #e83e8c; /* Güzel bir renk */
                text-align: center;
                margin-bottom: 30px;
            }
    
            ul {
                list-style-type: none;
                padding: 0;
            }
    
            li {
                border: 1px solid #dee2e6;
                border-radius: 5px;
                margin-bottom: 20px;
                padding: 20px;
                background-color: #fff;
                box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
                position: relative;
            }
    
            h4 {
                color: #e83e8c; /* Güzel bir renk */
                font-size: 18px;
            }
    
            form {
                display: flex;
                flex-direction: column;
                align-items: flex-end;
                margin-top: 10px;
            }
    
            .edit-form {
                margin-top: 10px;
            }
    
            label {
                color: #e83e8c; /* Güzel bir renk */
                font-weight: bold;
                font-size: 16px;
            }
    
            button {
                margin-top: 10px;
            }
    
            .archive-btn {
                background-color: #28a745;
                color: #fff;
                border: none;
                padding: 8px 15px;
                border-radius: 5px;
                cursor: pointer;
            }
    
            .archive-btn:hover {
                background-color: #218838;
            }
    
            /* Eklenen stil */
            .archived {
                display: none;
            }
    
            /* Stil eklemeleri */
            .btn-guncelle,
            .btn-sil {
                background-color: #e83e8c; /* Güzel bir renk */
                color: #fff;
                border: none;
                padding: 8px 15px;
                border-radius: 5px;
                cursor: pointer;
            }
    
            .btn-guncelle:hover,
            .btn-sil:hover {
                background-color: #c13167; /* Hover rengi */
            }
        </style>
    </head>
    
    <body>
        <div class="container">
            <h1>{{.PageTitle}}</h1>
            <ul>
                {{range .Tasks}}
                <li class="{{if .Archived.Valid}}{{if .Archived.Bool}}archived{{end}}{{end}}">
                    <h4>{{.Title}}</h4>
                    <p>{{.Description}}</p>
                    <form action="/edit" method="post" class="edit-form">
                        <input type="text" name="newText" placeholder="Yeni başlık" class="form-control">
                        <input type="hidden" name="id" value="{{.ID}}">
                        <button type="submit" class="btn btn-guncelle">&#8634; Güncelle</button>
                    </form>
                    <form action="/delete" method="post">
                        <input type="hidden" name="id" value="{{.ID}}">
                        <button type="submit" class="btn btn-sil">&#10006; Sil</button>
                    </form>
                    {{if .Archived.Valid}}
                        {{if .Archived.Bool}}
                            <button class="archive-btn" onclick="unarchiveTask({{.ID}})">Arşivden Çıkar</button>
                        {{else}}
                            <button class="archive-btn" onclick="archiveTask({{.ID}})">Arşivle</button>
                        {{end}}
                    {{else}}
                        <button class="archive-btn" onclick="archiveTask({{.ID}})">Arşivle</button>
                    {{end}}
                </li>
                {{end}}
            </ul>
            <form action="/add" method="post">
                <div class="form-group">
                    <label for="title">Başlık:</label>
                    <input type="text" id="title" name="title" required class="form-control">
                </div>
                <div class="form-group">
                    <label for="description">Açıklama:</label>
                    <textarea id="description" name="description" rows="3" class="form-control"></textarea>
                </div>
                <button type="submit" class="btn btn-primary">Ekle</button>
            </form>
        </div>
        <!-- Bootstrap JS ve Popper.js -->
        <script src="https://code.jquery.com/jquery-3.5.1.slim.min.js"></script>
        <script src="https://cdn.jsdelivr.net/npm/@popperjs/core@2.11.6/dist/umd/popper.min.js"></script>
        <script src="https://maxcdn.bootstrapcdn.com/bootstrap/4.5.2/js/bootstrap.min.js"></script>
        <script>
            function archiveTask(taskID) {
                console.log("Arşivle:", taskID);
                sendRequest('/archive', taskID);
            }
    
            function unarchiveTask(taskID) {
                console.log("Arşivden Çıkar:", taskID);
                sendRequest('/unarchive', taskID);
            }
    
            function sendRequest(url, taskID) {
                // AJAX kullanarak isteği gönder
                var xhr = new XMLHttpRequest();
                xhr.open('POST', url, true);
                xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
                xhr.onreadystatechange = function () {
                    if (xhr.readyState == 4 && xhr.status == 200) {
                        // Sayfayı yenile
                        location.reload();
                    }
                };
                xhr.send('id=' + taskID);
            }
        </script>
    </body>
    
    </html>
    `)
	if err != nil { //Hata var mı yok mu kontrol edilir.
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Page struct'ını doldur
	page := Page{
		PageTitle: "Yapılacak İşler", //Page struct'ındaki PageTitle alanına "Yapılacak İşler" değerini atar.
		Tasks:     tasks,             //getTasks fonksiyonu aracılığıyla alınan görevlerin  değerini atar.
	}

	// Template'e gönder
	err = tmpl.Execute(w, page) //tmpl adlı bir HTML şablonunu page adlı belirli bir veri bağlamıyla çalıştırmaya çalışması
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// addHandler, yeni görev eklemek için kullanılır.
func addHandler(w http.ResponseWriter, r *http.Request) {
	// Form verilerini al
	r.ParseForm()
	newTitle := r.Form.Get("title")
	newDescription := r.Form.Get("description")

	// Veritabanına yeni görevi ekle
	_, err := db.Exec("INSERT INTO tasks (title, description, archived) VALUES ($1, $2, $3)", newTitle, newDescription, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Ana sayfaya yönlendir
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// deleteHandler, görev silmek için kullanılır.
func deleteHandler(w http.ResponseWriter, r *http.Request) {
	// Form verilerini al
	r.ParseForm()
	taskID, err := strconv.Atoi(r.Form.Get("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Görevi veritabanından sil
	_, err = db.Exec("DELETE FROM tasks WHERE id = $1", taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Ana sayfaya yönlendir
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// editHandler, görevi güncellemek için kullanılır.
func editHandler(w http.ResponseWriter, r *http.Request) {
	// Form verilerini al
	r.ParseForm() //G elen HTTP isteğinin form verilerini analiz eder ve bu verileri `r.Form` haritasına yerleştirir.
	taskID, err := strconv.Atoi(r.Form.Get("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	newText := r.Form.Get("newText")

	// Görevi veritabanında güncelle
	_, err = db.Exec("UPDATE tasks SET title = $1 WHERE id = $2", newText, taskID) //exec dış komutlar için kullanıldı
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Ana sayfaya yönlendir
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// archiveHandler, görevi arşivler veya arşivden çıkarır.
func archiveHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	taskID, err := strconv.Atoi(r.Form.Get("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Görevin arşiv durumunu veritabanında güncelle
	_, err = db.Exec("UPDATE tasks SET archived = NOT archived WHERE id = $1", taskID) // Bu, bir SQL ifadesini veritabanında çalıştırmak için exec fonk.kullanıldı.
	if err != nil {                                                                    //Eğer hata nil değilse bir hata çıktısı alırız.
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Ana sayfaya yönlendir
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// getTasks, veritabanından tüm görevleri alır.
func getTasks() ([]Task, error) {
	// Veritabanından görevleri al
	rows, err := db.Query("SELECT id, title, description, archived FROM tasks")
	if err != nil {
		// Hata işleme
		fmt.Println("Görevleri alma hatası:", err)
		return nil, err
	}
	defer rows.Close()

	var tasks []Task // veritabanından alınan görev verilerini depolamak için kullanılacaktır.

	// Her bir satırı döngü içinde işle
	for rows.Next() { //Her döngü yenileme bir satırdaki verileri temsil eden `task` değişkenine yerleştirir.
		var task Task
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Archived) //rows.Scan` metodunun çağrılması sonucu oluşan hata durumunu kontrol eder.
		if err != nil {
			// Hata işleme
			fmt.Println("Satır okuma hatası:", err)
			return nil, err
		}

		// tasks slice'ına ekle
		tasks = append(tasks, task) //append` fonksiyonunun sonucunu `tasks` değişkenine atar. Çünkü `append` fonksiyonu, dilimde yapılan değişiklikleri geri döndürür ve bu değeri kullanarak dilimi güncelleme gerekir.
	}

	// rows hata kontrolü
	if err := rows.Err(); err != nil {
		// Hata işleme
		fmt.Println("Rows hatası:", err)
		return nil, err
	}

	return tasks, nil
}

func main() {
	// HTTP sunucusunu başlat
	http.HandleFunc("/", indexHandler)  //Ana sayfaya yapılan HTTP GET istekleri `indexHandler` adlı bir işlevle yönlendirilir.
	http.HandleFunc("/add", addHandler) //Bu, bir öğe eklemek için kullanılabilir.
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/edit", editHandler)
	http.HandleFunc("/archive", archiveHandler)

	http.ListenAndServe(":4656", nil)
}
