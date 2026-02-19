package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//go:embed static
var staticFS embed.FS

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"uniqueIndex"`
	Password string
}

type Run struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint
	Area       string
	Difficulty string
	Uniques    int
	Sets       int
	HRCount    int
	SessionSec int
	Timestamp  time.Time
}

type RuneDrop struct {
	ID    uint `gorm:"primaryKey"`
	RunID uint
	Rune  string
	Qty   int
}

var (
	db    *gorm.DB
	store = cookie.NewStore([]byte("d2r-secret-2026-change-me-in-prod"))

	areas = []string{
		"Countess (Hrabina)", "Radament", "Travincal Council", "Lower Kurast (LK)", "Mephisto",
		"Chaos Sanctuary", "Baal Waves", "Cow Level", "Pindleskin", "Nihlathak", "The Pit",
		"Ancient Tunnels", "Eldritch + Shenk", "Andariel", "Arcane Sanctuary", "Stony Tomb",
		"Arachnid Lair", "Maggot Lair", "Other",
	}

	difficulties = []string{"Normal", "Nightmare", "Hell"}

	highRunes = map[string]bool{
		"Um": true, "Mal": true, "Ist": true, "Gul": true, "Vex": true, "Ohm": true,
		"Lo": true, "Sur": true, "Ber": true, "Jah": true, "Cham": true, "Zod": true,
	}

	runeOrder = []string{
		"El", "Eld", "Tir", "Nef", "Eth", "Ith", "Tal", "Ral", "Ort", "Thul",
		"Amn", "Sol", "Shael", "Dol", "Hel", "Io", "Lum", "Ko", "Fal", "Lem",
		"Pul", "Um", "Mal", "Ist", "Gul", "Vex", "Ohm", "Lo", "Sur", "Ber",
		"Jah", "Cham", "Zod",
	}

	runeIcons = map[string]string{
		"El":   "https://static.wikia.nocookie.net/diablo/images/8/8f/El_Rune.png/revision/latest/scale-to-width-down/64",
		"Eld":  "https://static.wikia.nocookie.net/diablo/images/1/1d/Eld_Rune.png/revision/latest/scale-to-width-down/64",
		"Tir":  "https://static.wikia.nocookie.net/diablo/images/3/3f/Tir_Rune.png/revision/latest/scale-to-width-down/64",
		"Nef":  "https://static.wikia.nocookie.net/diablo/images/9/9e/Nef_Rune.png/revision/latest/scale-to-width-down/64",
		"Eth":  "https://static.wikia.nocookie.net/diablo/images/5/5f/Eth_Rune.png/revision/latest/scale-to-width-down/64",
		"Ith":  "https://static.wikia.nocookie.net/diablo/images/6/6e/Ith_Rune.png/revision/latest/scale-to-width-down/64",
		"Tal":  "https://static.wikia.nocookie.net/diablo/images/4/4f/Tal_Rune.png/revision/latest/scale-to-width-down/64",
		"Ral":  "https://static.wikia.nocookie.net/diablo/images/7/7f/Ral_Rune.png/revision/latest/scale-to-width-down/64",
		"Ort":  "https://static.wikia.nocookie.net/diablo/images/2/2f/Ort_Rune.png/revision/latest/scale-to-width-down/64",
		"Thul": "https://static.wikia.nocookie.net/diablo/images/0/0f/Thul_Rune.png/revision/latest/scale-to-width-down/64",
		"Amn":  "https://static.wikia.nocookie.net/diablo/images/9/9f/Amn_Rune.png/revision/latest/scale-to-width-down/64",
		"Sol":  "https://static.wikia.nocookie.net/diablo/images/5/5f/Sol_Rune.png/revision/latest/scale-to-width-down/64",
		"Shael":"https://static.wikia.nocookie.net/diablo/images/3/3f/Shael_Rune.png/revision/latest/scale-to-width-down/64",
		"Dol":  "https://static.wikia.nocookie.net/diablo/images/1/1f/Dol_Rune.png/revision/latest/scale-to-width-down/64",
		"Hel":  "https://static.wikia.nocookie.net/diablo/images/8/8f/Hel_Rune.png/revision/latest/scale-to-width-down/64",
		"Io":   "https://static.wikia.nocookie.net/diablo/images/2/2f/Io_Rune.png/revision/latest/scale-to-width-down/64",
		"Lum":  "https://static.wikia.nocookie.net/diablo/images/9/9f/Lum_Rune.png/revision/latest/scale-to-width-down/64",
		"Ko":   "https://static.wikia.nocookie.net/diablo/images/4/4f/Ko_Rune.png/revision/latest/scale-to-width-down/64",
		"Fal":  "https://static.wikia.nocookie.net/diablo/images/7/7f/Fal_Rune.png/revision/latest/scale-to-width-down/64",
		"Lem":  "https://static.wikia.nocookie.net/diablo/images/0/0f/Lem_Rune.png/revision/latest/scale-to-width-down/64",
		"Pul":  "https://static.wikia.nocookie.net/diablo/images/5/5f/Pul_Rune.png/revision/latest/scale-to-width-down/64",
		"Um":   "https://static.wikia.nocookie.net/diablo/images/3/3f/Um_Rune.png/revision/latest/scale-to-width-down/64",
		"Mal":  "https://static.wikia.nocookie.net/diablo/images/1/1f/Mal_Rune.png/revision/latest/scale-to-width-down/64",
		"Ist":  "https://static.wikia.nocookie.net/diablo/images/8/8f/Ist_Rune.png/revision/latest/scale-to-width-down/64",
		"Gul":  "https://static.wikia.nocookie.net/diablo/images/2/2f/Gul_Rune.png/revision/latest/scale-to-width-down/64",
		"Vex":  "https://static.wikia.nocookie.net/diablo/images/9/9f/Vex_Rune.png/revision/latest/scale-to-width-down/64",
		"Ohm":  "https://static.wikia.nocookie.net/diablo/images/4/4f/Ohm_Rune.png/revision/latest/scale-to-width-down/64",
		"Lo":   "https://static.wikia.nocookie.net/diablo/images/7/7f/Lo_Rune.png/revision/latest/scale-to-width-down/64",
		"Sur":  "https://static.wikia.nocookie.net/diablo/images/0/0f/Sur_Rune.png/revision/latest/scale-to-width-down/64",
		"Ber":  "https://static.wikia.nocookie.net/diablo/images/5/5f/Ber_Rune.png/revision/latest/scale-to-width-down/64",
		"Jah":  "https://static.wikia.nocookie.net/diablo/images/3/3f/Jah_Rune.png/revision/latest/scale-to-width-down/64",
		"Cham": "https://static.wikia.nocookie.net/diablo/images/1/1f/Cham_Rune.png/revision/latest/scale-to-width-down/64",
		"Zod":  "https://static.wikia.nocookie.net/diablo/images/8/8f/Zod_Rune.png/revision/latest/scale-to-width-down/64",
	}
)

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("d2r_tracker.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&User{}, &Run{}, &RuneDrop{})
}

func main() {
	initDB()

	r := gin.Default()
	r.Use(sessions.Sessions("d2rsession", store))

	// Statyczne pliki (opcjonalnie własne tło)
	r.StaticFS("/static", http.FS(staticFS))

	tmpl := template.Must(template.New("").Parse(d2rTemplate))
	r.SetHTMLTemplate(tmpl)

	r.GET("/", func(c *gin.Context) { c.Redirect(http.StatusFound, "/login") })

	// Auth routes
	r.GET("/login", loginPage)
	r.POST("/login", loginHandler)
	r.GET("/register", registerPage)
	r.POST("/register", registerHandler)
	r.GET("/logout", logoutHandler)

	// Protected
	protected := r.Group("/")
	protected.Use(authMiddleware())
	{
		protected.GET("/dashboard", dashboardHandler)
		protected.POST("/start-session", startSessionHandler)
		protected.POST("/log-run", logRunHandler)
		protected.GET("/leaderboard", leaderboardHandler)
		protected.GET("/my-stats", myStatsHandler)
	}

	log.Println("🚀 D2R Farm Tracker (D2R style) uruchomiony na http://localhost:8080")
	r.Run(":8080")
}

// ====================== MIDDLEWARE ======================
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		if session.Get("user_id") == nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

// ====================== HANDLERS ======================
func loginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "layout", gin.H{"Title": "Logowanie", "Content": loginHTML})
}

func registerPage(c *gin.Context) {
	c.HTML(http.StatusOK, "layout", gin.H{"Title": "Rejestracja", "Content": registerHTML})
}

func loginHandler(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	var user User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		c.HTML(http.StatusUnauthorized, "layout", gin.H{"Title": "Logowanie", "Content": loginHTML + `<p class="text-red-500 text-center mt-4">❌ Błędne dane</p>`})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		c.HTML(http.StatusUnauthorized, "layout", gin.H{"Title": "Logowanie", "Content": loginHTML + `<p class="text-red-500 text-center mt-4">❌ Błędne hasło</p>`})
		return
	}

	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Save()

	c.Redirect(http.StatusFound, "/dashboard")
}

func registerHandler(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	if username == "" || password == "" {
		c.HTML(http.StatusBadRequest, "layout", gin.H{"Title": "Rejestracja", "Content": registerHTML + `<p class="text-red-500">Wypełnij pola</p>`})
		return
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	db.Create(&User{Username: username, Password: string(hashed)})

	c.Redirect(http.StatusFound, "/login")
}

func logoutHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, "/login")
}

func dashboardHandler(c *gin.Context) {
	userID := sessions.Default(c).Get("user_id").(uint)

	var totalRuns, totalHR int64
	db.Model(&Run{}).Where("user_id = ?", userID).Count(&totalRuns)
	db.Model(&Run{}).Where("user_id = ?", userID).Select("COALESCE(SUM(hr_count),0)").Scan(&totalHR)

	content := fmt.Sprintf(`
		<div class="flex flex-col items-center">
			<div class="d2-panel w-full max-w-5xl">
				<div class="grid grid-cols-1 md:grid-cols-3 gap-8">
					<div class="text-center">
						<div class="text-6xl font-black text-amber-400 drop-shadow-d2">%d</div>
						<div class="text-xl tracking-widest text-red-400">RUNÓW</div>
					</div>
					<div class="text-center">
						<div class="text-6xl font-black text-emerald-400 drop-shadow-d2">%d</div>
						<div class="text-xl tracking-widest text-red-400">HIGH RUNES</div>
					</div>
					<div class="text-center">
						<button onclick="startSession()" 
								class="d2-btn px-12 py-6 text-2xl font-black tracking-widest">
							ROZPOCZNIJ SESJĘ
						</button>
						<div id="timer" class="mt-6 text-5xl font-mono text-amber-300">00:00:00</div>
					</div>
				</div>
			</div>

			<button onclick="showLogModal()" 
					class="mt-12 d2-btn-big text-3xl px-16 py-8">
				📜 ZALOGUJ NOWĄ RUNDĘ
			</button>

			<div class="mt-16 w-full max-w-5xl">
				<h2 class="text-4xl font-black text-center mb-8 text-amber-400">SIATKA RUN</h2>
				<div class="rune-grid">
					%s
				</div>
			</div>
		</div>

		<!-- MODAL -->
		<div id="logModal" class="hidden fixed inset-0 bg-black/90 flex items-center justify-center z-50">
			<div class="d2-panel max-w-4xl w-full mx-4">
				<div class="flex justify-between items-center border-b border-amber-400 pb-4">
					<h3 class="text-3xl font-black text-amber-400">LOG RUN – %s</h3>
					<button onclick="hideLogModal()" class="text-4xl text-red-400 hover:text-red-600">✕</button>
				</div>
				<form id="runForm" onsubmit="submitRun(event)" class="mt-8 space-y-8">
					<div class="grid grid-cols-2 gap-6">
						<select name="area" class="d2-input">%s</select>
						<select name="difficulty" class="d2-input">%s</select>
					</div>
					<div class="grid grid-cols-2 gap-6">
						<div>
							<label class="block text-amber-300 mb-2">Unikatów</label>
							<input type="number" name="uniques" value="0" class="d2-input w-full">
						</div>
						<div>
							<label class="block text-amber-300 mb-2">Zestawów</label>
							<input type="number" name="sets" value="0" class="d2-input w-full">
						</div>
					</div>

					<div>
						<label class="block text-amber-300 mb-3">Wybrane runy tej rundy</label>
						<div id="selectedRunes" class="flex flex-wrap gap-3 min-h-[60px]"></div>
					</div>

					<button type="submit" class="d2-btn-big w-full text-3xl py-6">✅ ZAPISZ RUNDĘ</button>
				</form>
			</div>
		</div>
	`, generateRuneGridHTML(), "HELL", generateAreaOptions(), generateDiffOptions())

	c.HTML(http.StatusOK, "layout", gin.H{
		"Title":   "Dashboard – D2R Farm Tracker",
		"Content": template.HTML(content),
	})
}

func generateRuneGridHTML() string {
	var sb strings.Builder
	for _, r := range runeOrder {
		icon := runeIcons[r]
		sb.WriteString(fmt.Sprintf(`
			<button onclick="addRuneToCurrent('%s')" 
					class="rune-btn group">
				<img src="%s" class="w-12 h-12 mx-auto group-hover:scale-110 transition">
				<div class="text-[10px] text-amber-400 mt-1">%s</div>
			</button>
		`, r, icon, r))
	}
	return sb.String()
}

func generateAreaOptions() string {
	var sb strings.Builder
	for _, a := range areas {
		sb.WriteString(fmt.Sprintf(`<option value="%s">%s</option>`, a, a))
	}
	return sb.String()
}

func generateDiffOptions() string {
	var sb strings.Builder
	for _, d := range difficulties {
		selected := ""
		if d == "Hell" {
			selected = ` selected`
		}
		sb.WriteString(fmt.Sprintf(`<option value="%s"%s>%s</option>`, d, selected, d))
	}
	return sb.String()
}

func startSessionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func logRunHandler(c *gin.Context) {
	userID := sessions.Default(c).Get("user_id").(uint)

	area := c.PostForm("area")
	diff := c.PostForm("difficulty")
	uniques, _ := strconv.Atoi(c.PostForm("uniques"))
	sets, _ := strconv.Atoi(c.PostForm("sets"))

	// Runy przychodzą jako JSON z JS (selectedRunes)
	var drops []struct {
		Rune string `json:"rune"`
		Qty  int    `json:"qty"`
	}
	jsonData := c.PostForm("runes")
	if jsonData != "" {
		json.Unmarshal([]byte(jsonData), &drops)
	}

	hrCount := 0
	for _, d := range drops {
		if highRunes[d.Rune] {
			hrCount += d.Qty
		}
	}

	run := Run{
		UserID:     userID,
		Area:       area,
		Difficulty: diff,
		Uniques:    uniques,
		Sets:       sets,
		HRCount:    hrCount,
		SessionSec: 0, // można rozszerzyć o timer
		Timestamp:  time.Now(),
	}
	db.Create(&run)

	for _, d := range drops {
		db.Create(&RuneDrop{RunID: run.ID, Rune: d.Rune, Qty: d.Qty})
	}

	c.JSON(http.StatusOK, gin.H{"status": "saved", "hr": hrCount})
}

func leaderboardHandler(c *gin.Context) {
	// Prosty leaderboard – top 10 HR
	type Leader struct {
		Username string
		TotalHR  int
		Runs     int
	}
	var leaders []Leader
	db.Raw(`
		SELECT u.username, SUM(r.hr_count) as total_hr, COUNT(r.id) as runs
		FROM runs r JOIN users u ON r.user_id = u.id
		GROUP BY u.id ORDER BY total_hr DESC LIMIT 10
	`).Scan(&leaders)

	// Generowanie HTML tabeli w D2R stylu
	var rows strings.Builder
	for i, l := range leaders {
		rows.WriteString(fmt.Sprintf(`
			<tr class="border-b border-amber-900 hover:bg-red-950/30">
				<td class="py-4 px-6 font-black">#%d</td>
				<td class="py-4 px-6">%s</td>
				<td class="py-4 px-6 text-emerald-400 text-right">%d HR</td>
				<td class="py-4 px-6 text-amber-400 text-right">%d runów</td>
			</tr>
		`, i+1, l.Username, l.TotalHR, l.Runs))
	}

	content := fmt.Sprintf(`<div class="d2-panel"><table class="w-full">%s</table></div>`, rows.String())

	c.HTML(http.StatusOK, "layout", gin.H{
		"Title":   "Leaderboard HR",
		"Content": template.HTML(content),
	})
}

func myStatsHandler(c *gin.Context) {
	// Wykres Chart.js – przykładowe dane (w realu pobierz z DB)
	content := `
		<div class="d2-panel">
			<canvas id="hrChart" class="w-full h-96"></canvas>
		</div>
		<script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
		<script>
			const ctx = document.getElementById('hrChart');
			new Chart(ctx, {
				type: 'line',
				data: {
					labels: ['Pon', 'Wt', 'Śr', 'Czw', 'Pt', 'Sob', 'Niedz'],
					datasets: [{
						label: 'High Runes / dzień',
						data: [3, 8, 12, 5, 15, 22, 9],
						borderColor: '#c9a14d',
						backgroundColor: 'rgba(201,161,77,0.2)',
						tension: 0.4,
						borderWidth: 4
					}]
				},
				options: { scales: { y: { grid: { color: '#4a2c0f' } } } }
			});
		</script>
	`

	c.HTML(http.StatusOK, "layout", gin.H{
		"Title":   "Moje statystyki",
		"Content": template.HTML(content),
	})
}

// ====================== TEMPLATES ======================
var d2rTemplate = `
<!DOCTYPE html>
<html lang="pl">
<head>
	<meta charset="UTF-8">
	<title>{{.Title}} - D2R Farm Tracker</title>
	<script src="https://cdn.tailwindcss.com"></script>
	<script src="https://unpkg.com/htmx.org@1.9.12"></script>
	<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Creepster&family=UnifrakturMaguntia&display=swap">
	<style>
		@layer base {
			body {
				background: radial-gradient(circle at center, #1c1208 0%, #0a0503 100%);
				background-image: url('https://i.imgur.com/2vL9k8P.png'); /* subtelne D2 texture */
				background-size: cover;
				color: #e8d5a3;
				font-family: system-ui, sans-serif;
			}
			.d2-panel {
				background: linear-gradient(#2c1f14, #1a1109);
				border: 6px solid #d4af37;
				box-shadow: 0 0 30px #8b0000, inset 0 0 20px rgba(212,175,55,0.3);
				padding: 2rem;
				border-radius: 4px;
			}
			.d2-btn {
				background: linear-gradient(to bottom, #8b5a2b, #5c3a1a);
				border: 4px solid #d4af37;
				color: #ffd700;
				text-shadow: 2px 2px 0 #4a1c1c;
				transition: all 0.2s;
			}
			.d2-btn:hover {
				background: linear-gradient(to bottom, #b38b4d, #8b5a2b);
				transform: scale(1.05);
			}
			.d2-btn-big {
				background: linear-gradient(to bottom, #b71c1c, #7f0000);
				border: 6px solid #ffd700;
				color: #fff;
				text-shadow: 3px 3px 0 #000;
			}
			.d2-input {
				background: #1a1109;
				border: 3px solid #d4af37;
				color: #e8d5a3;
				padding: 12px;
				font-size: 1.1rem;
			}
			.rune-btn {
				background: #1c1208;
				border: 3px solid #4a2c0f;
				transition: all 0.15s;
			}
			.rune-btn:hover {
				border-color: #ffd700;
				transform: translateY(-4px) scale(1.1);
				box-shadow: 0 0 15px #ffd700;
			}
			.drop-shadow-d2 {
				text-shadow: 3px 3px 0 #4a1c1c, -2px -2px 0 #4a1c1c;
			}
			.rune-grid {
				display: grid;
				grid-template-columns: repeat(auto-fit, minmax(80px, 1fr));
				gap: 12px;
			}
		}
	</style>
</head>
<body class="min-h-screen">
	<div class="max-w-screen-xl mx-auto p-6">
		<header class="flex justify-between items-center mb-12 border-b border-amber-400 pb-6">
			<div class="flex items-center gap-4">
				<div class="text-6xl font-black tracking-[6px] text-[#c9a14d] drop-shadow-d2">DIABLO</div>
				<div class="text-5xl text-red-500 font-black tracking-widest">II</div>
			</div>
			<div class="flex gap-6 text-xl">
				<a href="/dashboard" class="hover:text-amber-400 transition">DASHBOARD</a>
				<a href="/leaderboard" class="hover:text-amber-400 transition">LEADERBOARD</a>
				<a href="/my-stats" class="hover:text-amber-400 transition">MOJE STATY</a>
				<a href="/logout" class="text-red-400 hover:text-red-600">WYLOGUJ</a>
			</div>
		</header>

		{{.Content}}
	</div>

	<script>
		let currentRunes = [];

		function addRuneToCurrent(rune) {
			const qty = prompt("Ile sztuk " + rune + "?", "1");
			if (!qty) return;
			const num = parseInt(qty);
			if (num > 0) {
				currentRunes.push({rune: rune, qty: num});
				renderSelectedRunes();
			}
		}

		function renderSelectedRunes() {
			const container = document.getElementById("selectedRunes");
			container.innerHTML = currentRunes.map((item, i) => `
				<div onclick="removeRune(${i})" class="flex items-center gap-3 bg-zinc-900 border border-amber-400 px-4 py-2 rounded cursor-pointer hover:bg-red-900">
					<img src="${window.runeIcons[item.rune] || ''}" class="w-8 h-8">
					<span>${item.rune} × ${item.qty}</span>
				</div>
			`).join('');
		}

		function removeRune(i) {
			currentRunes.splice(i, 1);
			renderSelectedRunes();
		}

		function showLogModal() {
			currentRunes = [];
			renderSelectedRunes();
			document.getElementById("logModal").classList.remove("hidden");
		}

		function hideLogModal() {
			document.getElementById("logModal").classList.add("hidden");
		}

		function submitRun(e) {
			e.preventDefault();
			const formData = new FormData(e.target);
			formData.append("runes", JSON.stringify(currentRunes));

			fetch("/log-run", {method: "POST", body: formData})
				.then(r => r.json())
				.then(data => {
					alert("✅ Runda zapisana! HR: " + data.hr);
					hideLogModal();
					location.reload();
				});
		}

		// Timer sesji (prosty JS)
		let timerInterval;
		function startSession() {
			let seconds = 0;
			clearInterval(timerInterval);
			timerInterval = setInterval(() => {
				seconds++;
				const h = Math.floor(seconds / 3600).toString().padStart(2,'0');
				const m = Math.floor((seconds % 3600) / 60).toString().padStart(2,'0');
				const s = (seconds % 60).toString().padStart(2,'0');
				document.getElementById("timer").textContent = h + ":" + m + ":" + s;
			}, 1000);
			alert("⏳ Sesja rozpoczęta! Loguj runy kiedy chcesz.");
		}

		// Globalne ikony dla JS
		window.runeIcons = {{.RuneIconsJSON}};
	</script>
</body>
</html>
`

var loginHTML = `
<div class="max-w-md mx-auto mt-32 d2-panel">
	<h1 class="text-5xl font-black text-center mb-10 text-amber-400">LOGOWANIE</h1>
	<form method="POST" action="/login" class="space-y-6">
		<input name="username" placeholder="Nazwa bohatera" required class="d2-input w-full">
		<input name="password" type="password" placeholder="Hasło" required class="d2-input w-full">
		<button type="submit" class="d2-btn-big w-full py-6 text-2xl">WEJDŹ DO SANKTUARIUM</button>
	</form>
	<p class="text-center mt-8"><a href="/register" class="text-amber-400 hover:underline">Nowy bohater? Zarejestruj się</a></p>
</div>
`

var registerHTML = `
<div class="max-w-md mx-auto mt-32 d2-panel">
	<h1 class="text-5xl font-black text-center mb-10 text-amber-400">NOWY BOHATER</h1>
	<form method="POST" action="/register" class="space-y-6">
		<input name="username" placeholder="Nazwa bohatera" required class="d2-input w-full">
		<input name="password" type="password" placeholder="Hasło" required class="d2-input w-full">
		<button type="submit" class="d2-btn-big w-full py-6 text-2xl">STWÓRZ POSTAĆ</button>
	</form>
</div>
`
