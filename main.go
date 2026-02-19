package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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
	store = cookie.NewStore([]byte("d2r-secret-render-2026"))

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
		"Shael": "https://static.wikia.nocookie.net/diablo/images/3/3f/Shael_Rune.png/revision/latest/scale-to-width-down/64",
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
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		db, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{})
		log.Println("✅ Połączono z PostgreSQL (Render.com)")
	} else {
		db, err = gorm.Open(sqlite.Open("d2r_tracker.db"), &gorm.Config{})
		log.Println("✅ Używam lokalnego SQLite")
	}
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&User{}, &Run{}, &RuneDrop{})
}

func main() {
	initDB()
	r := gin.Default()
	r.Use(sessions.Sessions("d2rsession", store))

	tmpl := template.Must(template.New("").Parse(d2rTemplate))
	r.SetHTMLTemplate(tmpl)

	r.GET("/", func(c *gin.Context) { c.Redirect(http.StatusFound, "/login") })

	r.GET("/login", loginPage)
	r.POST("/login", loginHandler)
	r.GET("/register", registerPage)
	r.POST("/register", registerHandler)
	r.GET("/logout", logoutHandler)

	protected := r.Group("/")
	protected.Use(authMiddleware())
	{
		protected.GET("/dashboard", dashboardHandler)
		protected.POST("/log-run", logRunHandler)
		protected.GET("/leaderboard", leaderboardHandler)
		protected.GET("/my-stats", myStatsHandler)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 D2R Farm Tracker v2.1 uruchomiony na http://localhost:%s", port)
	r.Run(":" + port)
}

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

// ==================== AUTH ====================
func loginPage(c *gin.Context) { c.HTML(http.StatusOK, "layout", gin.H{"Title": "Logowanie", "Content": loginHTML}) }
func registerPage(c *gin.Context) { c.HTML(http.StatusOK, "layout", gin.H{"Title": "Rejestracja", "Content": registerHTML}) }

func loginHandler(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	var user User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		c.HTML(http.StatusUnauthorized, "layout", gin.H{"Title": "Logowanie", "Content": loginHTML + `<p class="text-red-500 text-center mt-6">❌ Błędne dane</p>`})
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

// ==================== DASHBOARD ====================
func dashboardHandler(c *gin.Context) {
	userID := sessions.Default(c).Get("user_id").(uint)
	var totalRuns, totalHR, totalUniques, totalSets int64
	db.Model(&Run{}).Where("user_id = ?", userID).Count(&totalRuns)
	db.Model(&Run{}).Where("user_id = ?", userID).Select("COALESCE(SUM(hr_count),0)").Scan(&totalHR)
	db.Model(&Run{}).Where("user_id = ?", userID).Select("COALESCE(SUM(uniques),0)").Scan(&totalUniques)
	db.Model(&Run{}).Where("user_id = ?", userID).Select("COALESCE(SUM(sets),0)").Scan(&totalSets)

	avgHR := 0.0
	if totalRuns > 0 {
		avgHR = float64(totalHR) / float64(totalRuns)
	}

	content := fmt.Sprintf(`
		<div class="flex flex-col items-center">
			<div class="d2-panel w-full max-w-6xl mx-auto">
				<div class="grid grid-cols-2 md:grid-cols-4 gap-8 text-center">
					<div><div class="text-7xl font-black text-amber-400">%d</div><div class="text-xl tracking-widest">RUNÓW</div></div>
					<div><div class="text-7xl font-black text-emerald-400">%d</div><div class="text-xl tracking-widest">HIGH RUNES</div></div>
					<div><div class="text-7xl font-black text-amber-400">%.2f</div><div class="text-xl tracking-widest">HR / RUN</div></div>
					<div><button onclick="startSession()" class="d2-btn px-10 py-6 text-2xl font-black tracking-widest">ROZPOCZNIJ SESJĘ</button><div id="timer" class="mt-6 text-5xl font-mono text-amber-300">00:00:00</div></div>
				</div>
			</div>

			<button onclick="showLogModal()" class="mt-12 d2-btn-big text-4xl px-24 py-10 font-black">📜 ZALOGUJ NOWĄ RUNDĘ</button>

			<div class="mt-16 w-full max-w-6xl grid grid-cols-3 gap-8">
				<div class="d2-panel text-center"><div class="text-6xl font-black">%d</div><div class="text-amber-300">UNIKATÓW</div></div>
				<div class="d2-panel text-center"><div class="text-6xl font-black">%d</div><div class="text-amber-300">ZESTAWÓW</div></div>
				<div class="d2-panel text-center"><div class="text-6xl font-black text-emerald-400">%.1f%%</div><div class="text-amber-300">EFFICIENCY</div></div>
			</div>
		</div>

		<div id="logModal" class="hidden fixed inset-0 bg-black/95 flex items-center justify-center z-50">
			<div class="d2-panel w-full max-w-4xl mx-4">
				<div class="flex justify-between border-b border-amber-400 pb-4">
					<h3 class="text-4xl font-black text-amber-400">LOG RUN</h3>
					<button onclick="hideLogModal()" class="text-5xl text-red-400">✕</button>
				</div>
				<form id="runForm" onsubmit="submitRun(event)" class="mt-8 space-y-8">
					<div class="grid grid-cols-2 gap-6">
						<select name="area" class="d2-input">%s</select>
						<select name="difficulty" class="d2-input">%s</select>
					</div>
					<div class="grid grid-cols-2 gap-6">
						<div><label class="block text-amber-300">Unikatów</label><input type="number" name="uniques" value="0" class="d2-input w-full"></div>
						<div><label class="block text-amber-300">Zestawów</label><input type="number" name="sets" value="0" class="d2-input w-full"></div>
					</div>
					<div>
						<label class="block text-amber-300 mb-3">Runy (klikaj na siatce poniżej)</label>
						<div id="selectedRunes" class="flex flex-wrap gap-3 min-h-[70px]"></div>
					</div>
					<button type="submit" class="d2-btn-big w-full py-8 text-3xl">✅ ZAPISZ RUNDĘ</button>
				</form>
			</div>
		</div>

		<div class="mt-16 w-full max-w-6xl">
			<h2 class="text-4xl font-black text-center mb-8 text-amber-400">SIATKA RUN</h2>
			<div class="rune-grid">%s</div>
		</div>
	`, generateAreaOptions(), generateDiffOptions(), generateRuneGridHTML())

	c.HTML(http.StatusOK, "layout", gin.H{"Title": "Dashboard – D2R Farm Tracker", "Content": template.HTML(content)})
}

func generateAreaOptions() string { var s strings.Builder; for _, a := range areas { s.WriteString(fmt.Sprintf(`<option value="%s">%s</option>`, a, a)) }; return s.String() }
func generateDiffOptions() string { var s strings.Builder; for _, d := range difficulties { sel := ""; if d == "Hell" { sel = ` selected` }; s.WriteString(fmt.Sprintf(`<option value="%s"%s>%s</option>`, d, sel, d)) }; return s.String() }
func generateRuneGridHTML() string {
	var sb strings.Builder
	for _, r := range runeOrder {
		sb.WriteString(fmt.Sprintf(`<button onclick="addRuneToCurrent('%s')" class="rune-btn"><img src="%s" class="w-14 h-14"><div class="text-xs mt-1">%s</div></button>`, r, runeIcons[r], r))
	}
	return sb.String()
}

// ==================== LOG RUN ====================
func logRunHandler(c *gin.Context) {
	userID := sessions.Default(c).Get("user_id").(uint)
	area := c.PostForm("area")
	diff := c.PostForm("difficulty")
	uniques, _ := strconv.Atoi(c.PostForm("uniques"))
	sets, _ := strconv.Atoi(c.PostForm("sets"))

	var drops []struct{ Rune string `json:"rune"`; Qty int `json:"qty"` }
	if j := c.PostForm("runes"); j != "" {
		json.Unmarshal([]byte(j), &drops)
	}

	hr := 0
	for _, d := range drops {
		if highRunes[d.Rune] {
			hr += d.Qty
		}
	}

	run := Run{UserID: userID, Area: area, Difficulty: diff, Uniques: uniques, Sets: sets, HRCount: hr, Timestamp: time.Now()}
	db.Create(&run)
	for _, d := range drops {
		db.Create(&RuneDrop{RunID: run.ID, Rune: d.Rune, Qty: d.Qty})
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "hr": hr})
}

// ==================== LEADERBOARD ====================
func leaderboardHandler(c *gin.Context) {
	type L struct {
		Username string
		TotalHR  int
		Runs     int
		AvgHR    float64
	}
	var leaders []L
	db.Raw(`SELECT u.username, SUM(r.hr_count) AS total_hr, COUNT(*) AS runs, AVG(r.hr_count) AS avg_hr 
			FROM runs r JOIN users u ON r.user_id = u.id 
			GROUP BY u.id ORDER BY total_hr DESC LIMIT 15`).Scan(&leaders)

	var rows strings.Builder
	for i, l := range leaders {
		rows.WriteString(fmt.Sprintf(`<tr class="border-b border-amber-900"><td class="py-4 px-6 font-black">#%d</td><td>%s</td><td class="text-emerald-400">%d HR</td><td>%d runów</td><td>%.2f HR/run</td></tr>`, i+1, l.Username, l.TotalHR, l.Runs, l.AvgHR))
	}

	content := fmt.Sprintf(`<div class="d2-panel"><h2 class="text-4xl font-black mb-8 text-center">🏆 LEADERBOARD HR</h2><table class="w-full">%s</table></div>`, rows.String())
	c.HTML(http.StatusOK, "layout", gin.H{"Title": "Leaderboard", "Content": template.HTML(content)})
}

// ==================== MY STATS ====================
func myStatsHandler(c *gin.Context) {
	content := `<div class="d2-panel"><canvas id="chart" class="w-full h-96"></canvas></div>
	<script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
	<script>
		new Chart(document.getElementById("chart"), {type:"line", data:{labels:["Pon","Wt","Śr","Czw","Pt","Sob","Niedz"], datasets:[{label:"HR", data:[4,11,7,18,25,32,14], borderColor:"#c9a14d", tension:0.4}]}, options:{scales:{y:{grid:{color:"#4a2c0f"}}}}});
	</script>`
	c.HTML(http.StatusOK, "layout", gin.H{"Title": "Moje statystyki", "Content": template.HTML(content)})
}

// ==================== TEMPLATES ====================
var d2rTemplate = `
<!DOCTYPE html>
<html lang="pl">
<head>
	<meta charset="UTF-8">
	<title>{{.Title}}</title>
	<script src="https://cdn.tailwindcss.com"></script>
	<style>
		body { background: radial-gradient(#1c1208, #0a0503); color: #e8d5a3; font-family: system-ui; }
		.d2-panel { background: linear-gradient(#2c1f14,#1a1109); border: 8px solid #d4af37; box-shadow: 0 0 40px #8b0000; padding: 2rem; border-radius: 4px; }
		.d2-btn { background: linear-gradient(#8b5a2b,#5c3a1a); border: 4px solid #d4af37; color: #ffd700; padding: 12px 24px; font-weight: 900; }
		.d2-btn-big { background: linear-gradient(#b71c1c,#7f0000); border: 6px solid #ffd700; color: #fff; padding: 20px 40px; font-size: 1.5rem; font-weight: 900; }
		.rune-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(90px,1fr)); gap: 14px; }
		.rune-btn { background:#1c1208; border:4px solid #4a2c0f; transition:all .2s; padding: 8px; text-align: center; }
		.rune-btn:hover { border-color:#ffd700; transform:scale(1.12); }
	</style>
</head>
<body class="min-h-screen">
	<div class="max-w-screen-2xl mx-auto p-8">
		<header class="flex justify-between items-center mb-12 border-b border-amber-400 pb-6">
			<div class="flex items-center gap-6">
				<div class="text-7xl font-black text-[#c9a14d] tracking-widest">DIABLO</div>
				<div class="text-6xl text-red-600 font-black">II</div>
			</div>
			<div class="flex gap-10 text-2xl">
				<a href="/dashboard" class="hover:text-amber-400">Dashboard</a>
				<a href="/leaderboard" class="hover:text-amber-400">Leaderboard</a>
				<a href="/my-stats" class="hover:text-amber-400">Moje staty</a>
				<a href="/logout" class="text-red-500">Wyloguj</a>
			</div>
		</header>
		{{.Content}}
	</div>

	<script>
		let currentRunes = [];
		function addRuneToCurrent(r) {
			let qty = prompt("Ile sztuk "+r+"?", "1");
			if(qty && parseInt(qty)>0) {
				currentRunes.push({rune:r, qty:parseInt(qty)});
				renderSelected();
			}
		}
		function renderSelected() {
			let html = currentRunes.map((item,i) => `<div onclick="removeRune(\( {i})" class="flex items-center gap-3 bg-zinc-900 border border-amber-400 px-5 py-3 rounded cursor-pointer hover:bg-red-900"> \){item.rune} × ${item.qty}</div>`).join('');
			document.getElementById("selectedRunes").innerHTML = html || "Kliknij runy powyżej...";
		}
		function removeRune(i) { currentRunes.splice(i,1); renderSelected(); }
		function showLogModal() { currentRunes = []; renderSelected(); document.getElementById("logModal").classList.remove("hidden"); }
		function hideLogModal() { document.getElementById("logModal").classList.add("hidden"); }
		function submitRun(e) {
			e.preventDefault();
			const form = new FormData(e.target);
			form.append("runes", JSON.stringify(currentRunes));
			fetch("/log-run", {method:"POST", body:form}).then(r=>r.json()).then(d=>{
				alert("✅ Zapisano! HR: "+d.hr);
				hideLogModal();
				location.reload();
			});
		}
		let seconds = 0, timer;
		function startSession() {
			clearInterval(timer);
			seconds = 0;
			timer = setInterval(()=>{ seconds++; let h=Math.floor(seconds/3600),m=Math.floor((seconds%3600)/60),s=seconds%60; document.getElementById("timer").textContent = h.toString().padStart(2,'0')+":"+m.toString().padStart(2,'0')+":"+s.toString().padStart(2,'0'); },1000);
			alert("⏳ Sesja rozpoczęta!");
		}
	</script>
</body>
</html>
`

var loginHTML = `
<div class="max-w-md mx-auto mt-32 d2-panel p-12">
	<h1 class="text-5xl font-black text-center mb-12 text-amber-400">LOGOWANIE</h1>
	<form method="POST" action="/login" class="space-y-8">
		<input name="username" placeholder="Nazwa bohatera" required class="d2-input w-full p-5 text-xl">
		<input name="password" type="password" placeholder="Hasło" required class="d2-input w-full p-5 text-xl">
		<button type="submit" class="d2-btn-big w-full py-8 text-3xl">WEJDŹ DO SANKTUARIUM</button>
	</form>
	<p class="text-center mt-10"><a href="/register" class="text-amber-400 text-xl">Nowy bohater? Zarejestruj się</a></p>
</div>
`

var registerHTML = `
<div class="max-w-md mx-auto mt-32 d2-panel p-12">
	<h1 class="text-5xl font-black text-center mb-12 text-amber-400">STWÓRZ BOHATERA</h1>
	<form method="POST" action="/register" class="space-y-8">
		<input name="username" placeholder="Nazwa bohatera" required class="d2-input w-full p-5 text-xl">
		<input name="password" type="password" placeholder="Hasło" required class="d2-input w-full p-5 text-xl">
		<button type="submit" class="d2-btn-big w-full py-8 text-3xl">STWÓRZ POSTAĆ</button>
	</form>
</div>
`
