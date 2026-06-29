package profile

// Profile is the agent's value system: what a good job looks like, plus a
// resume summary the LLM uses to write tailored applications. Edit this file
// to retune everything — no other code changes needed.
var (
	Name    = "Emmanuel Ngulube (Dan)"
	BasedIn = "Abuja, Nigeria"
	Email   = "emmanuelngulube292@gmail.com"

	// Roles wanted; hits in the job title weigh heavily.
	Roles = []string{
		"frontend", "front end", "front-end", "mobile", "react native",
		"react developer", "software engineer", "full stack", "fullstack",
	}

	// Stack; hits anywhere add relevance.
	Stack = []string{
		"react native", "expo", "react", "next.js", "nextjs", "typescript",
		"javascript", "nativewind", "tailwind", "go", "golang", "chi",
	}

	// Seniority Dan can realistically land; hits add a bonus.
	GoodSeniority = []string{
		"intern", "internship", "junior", "entry", "graduate", "new grad",
		"associate", "sde 1", "sde-1", "sde i", "software engineer i", "early career",
	}

	// Seniority that's a stretch; hits subtract.
	HardSeniority = []string{
		"senior", "staff", "principal", "lead", "head of", "director", "vp ",
	}

	// Signals that the role is US-work-authorization-only (usually a blocker).
	USOnlySignals = []string{
		"us only", "u.s. only", "united states only", "must be based in the us",
		"must be located in the united states", "must reside in the us",
		"authorized to work in the us", "us work authorization", "us citizen",
		"based in the us", "us-based", "usa only",
	}

	// Signals Dan CAN apply from anywhere / Africa.
	GlobalSignals = []string{
		"worldwide", "global remote", "remote anywhere", "anywhere in the world",
		"work from anywhere", "fully remote", "africa", "nigeria", "emea",
		"any timezone", "global",
	}

	// Resume summary the LLM grounds cover letters in. Keep it true.
	ResumeSummary = `Emmanuel "Dan" Ngulube — Frontend & Mobile Engineer based in Abuja, Nigeria.
Final-year Computer Science student at Veritas University Abuja. Co-founder of Ogbiogbi Limited (freelance studio); completed internships at Olotusquare and Jadaad Technologies.
Core stack: React Native, Expo, React, Next.js, TypeScript, NativeWind/Tailwind. Backend in Go (Chi, REST APIs) with PostgreSQL and MongoDB — actively going deeper on Go and system design.
Selected work: Proxi (location-based reminders, React Native + Go geofencing); ThreatIQ (community safety intelligence, Next.js + Go + maps + AI risk scoring); Reflecta (mood-tracking mobile app, React Native + Go); BizCrew AI (AI SaaS for small businesses, Next.js + Supabase + LLMs).
Portfolio: dantech-xi.vercel.app · GitHub: github.com/Justdan111 · LinkedIn: linkedin.com/in/ngulube-emmanuel
Looking for: remote frontend or mobile engineering roles at funded/early-stage startups that ship fast.`
)
