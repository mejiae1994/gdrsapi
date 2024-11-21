package steamrating

import (
	"encoding/json"
	"fmt"
	"gdrsapi/external/cloudflare"
	"gdrsapi/external/gemini"
	"gdrsapi/external/gsheets"
	"gdrsapi/pkg/logger"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

var genreToTags = map[string][]string{
	"action": {
		"action", "hack and slash", "hack-and-slash", "beat 'em up", "brawler",
		"fighting", "martial arts", "third-person", "melee combat", "spectacle fighter",
		"character action game", "sword fighting", "gun fu", "bullet time", "combo-based",
		"fast-paced", "reflex-based", "hand-to-hand combat", "weapon-based fighter",
		"arena combat", "action-adventure", "parkour", "quick time events", "qte",
		"cinematic action", "stealth action", "assassin", "ninja", "samurai",
	},
	"adventure": {
		"adventure", "exploration", "story-rich", "point-and-click", "narrative",
		"choice matter", "interactive fiction", "text-based", "walking simulator",
		"visual novel", "puzzle-adventure", "hidden object", "escape room", "mystery",
		"detective", "thriller", "horror adventure", "survival horror", "psychological horror",
		"adventure rpg", "action-adventure", "open world adventure", "historical adventure",
		"sci-fi adventure", "fantasy adventure", "episodic", "story-driven", "branching narrative",
		"multiple endings", "time travel", "archaeology", "treasure hunting",
	},
	"strategy": {
		"strategy", "turn-based", "turn based", "real-time strategy", "rts", "4x",
		"grand strategy", "tower defense", "auto battler", "tactical", "wargame",
		"card game", "deck-building", "moba", "base-building", "city-builder",
		"resource management", "economy", "diplomacy", "political", "historical",
		"military", "battle simulator", "tactics", "squad-based tactics", "hero collector",
		"multiplayer online battle arena", "tower offense", "defense", "automation",
		"programming", "hacking", "cyberpunk", "space strategy", "naval", "trading",
	},
	"rpg": {
		"rpg", "role playing", "role-playing", "action rpg", "jrpg", "crpg",
		"party-based", "turn-based rpg", "open world", "character customization",
		"dungeon crawler", "building", "farming", "western rpg", "sandbox rpg",
		"tactical rpg", "roguelike rpg", "action-adventure rpg", "mmorpg", "online rpg",
		"story-rich rpg", "choice matter", "multiple endings", "class-based",
		"skill tree", "leveling system", "loot-based", "crafting", "alchemy",
		"magic system", "fantasy rpg", "sci-fi rpg", "post-apocalyptic rpg",
		"cyberpunk rpg", "steampunk rpg", "historical rpg", "medieval rpg",
	},
	"simulation": {
		"simulation", "life sim", "farm sim", "management", "tycoon", "business sim",
		"dating sim", "social sim", "space sim", "flight sim", "train sim", "truck sim",
		"cooking sim", "city-builder", "pet sim", "animal sim", "medical sim", "surgery sim",
		"sports management", "political sim", "war sim", "ecosystem sim", "physics sim",
		"vehicle sim", "driving sim", "racing sim", "sailing sim", "submarine sim",
		"economy sim", "government sim", "colony sim", "survival sim", "crafting sim",
		"building sim", "automation sim", "factory sim", "agriculture sim", "biology sim",
	},
	"sports": {
		"sports", "football", "soccer", "basketball", "baseball", "golf", "tennis",
		"wrestling", "extreme sports", "team sports", "sports management", "olympics",
		"hockey", "ice hockey", "volleyball", "beach volleyball", "cricket", "rugby",
		"american football", "boxing", "mma", "martial arts", "skateboarding", "snowboarding",
		"skiing", "surfing", "bmx", "cycling", "athletics", "track and field", "swimming",
		"diving", "gymnastics", "billiards", "pool", "snooker", "darts", "bowling",
		"table tennis", "ping pong", "badminton", "lacrosse", "water sports",
	},
	"racing": {
		"racing", "car racing", "motorcycle racing", "offroad", "racing sim",
		"kart racing", "arcade racing", "rally", "drag racing", "motocross",
		"formula racing", "stock car racing", "street racing", "futuristic racing",
		"bike racing", "boat racing", "jet ski racing", "hovercraft racing", "racing management",
		"time attack", "drift racing", "demolition derby", "truck racing", "buggy racing",
		"atv racing", "snowmobile racing", "racing rpg", "open world racing", "racing strategy",
	},
	"puzzle": {
		"puzzle", "logic", "physics puzzle", "match-3", "hidden object", "escape room",
		"jigsaw puzzle", "sudoku", "word game", "programming puzzle", "block-pushing puzzle",
		"sliding puzzle", "pattern recognition", "memory puzzle", "math puzzle", "riddle",
		"maze", "sokoban", "bridge-building", "contraption-builder", "puzzle platformer",
		"puzzle-adventure", "casual puzzle", "bubble shooter", "tile-matching", "tangram",
		"crossword", "logic grid", "picross", "nonogram", "cryptogram", "anagram",
		"spatial reasoning", "color matching", "connect the dots", "pipe connecting",
	},
	"arcade": {
		"arcade", "retro", "classic", "score attack", "endless runner", "rhythm",
		"music game", "pinball", "breakout", "shoot 'em up", "bullet hell", "side-scroller",
		"beat 'em up", "fighting game", "light gun", "rail shooter", "maze game", "platformer",
		"twin-stick shooter", "fixed shooter", "puzzle bobble", "tetris-like", "pong-like",
		"pac-man-like", "space invaders-like", "galaga-like", "donkey kong-like", "frogger-like",
		"centipede-like", "asteroids-like", "defender-like", "joust-like", "qbert-like",
		"dig dug-like", "bubble bobble-like", "rampage-like", "gauntlet-like",
	},
	"platformer": {
		"platformer", "2D platformer", "3D platformer", "metroidvania", "run and gun",
		"precision platformer", "puzzle platformer", "action platformer", "cinematic platformer",
		"physics-based platformer", "endless platformer", "roguelike platformer", "auto-runner",
		"side-scroller", "exploration platformer", "collectathon", "mascot platformer",
		"parkour platformer", "stealth platformer", "speedrun platformer", "hardcore platformer",
		"platformer shooter", "platformer rpg", "co-op platformer", "competitive platformer",
		"wall-jumping", "double-jump", "grappling hook", "swinging mechanics",
	},
	"shooter": {
		"shooter", "fps", "first-person shooter", "third-person shooter", "shmup", "shoot 'em up",
		"bullet hell", "tactical shooter", "arena shooter", "on-rails shooter", "battle royale",
		"hero shooter", "looter shooter", "cover shooter", "team-based shooter", "class-based shooter",
		"mil-sim", "arcade shooter", "vehicular combat", "space shooter", "zombie shooter",
		"survival shooter", "co-op shooter", "twin-stick shooter", "top-down shooter",
		"isometric shooter", "stealth shooter", "time-manipulation shooter", "retro shooter",
		"physics-based shooter", "sci-fi shooter", "realistic shooter", "western shooter",
	},
	"visual novel": {
		"visual novel", "otome", "kinetic novel", "dating sim", "choice matter", "multiple endings",
		"romance", "interactive fiction", "text-based", "story-rich", "branching narrative",
		"character-driven", "dialogue-heavy", "slice of life", "mystery visual novel",
		"horror visual novel", "sci-fi visual novel", "fantasy visual novel", "historical visual novel",
		"psychological", "drama", "comedy", "thriller", "supernatural", "school life", "coming of age",
		"adult", "all-ages", "boys' love", "girls' love", "harem", "reverse harem", "episodic",
	},
	"tabletop": {
		"tabletop", "board game", "card game", "dice", "chess", "gambling", "tabletop rpg",
		"collectible card game", "deck-building", "miniatures", "tile-placement", "worker placement",
		"area control", "strategy board game", "party game", "social deduction", "hidden role",
		"cooperative board game", "legacy board game", "eurogame", "ameritrash", "abstract strategy",
		"wargame", "roll and write", "auction", "drafting", "push your luck", "real-time",
		"dexterity", "memory", "word game", "trivia", "escape room game", "dungeon crawler",
	},
	"roguelike": {
		"roguelike", "roguelite", "rogue-like", "rogue-lite", "procedural generation", "permadeath",
		"dungeon crawler", "run-based", "randomized", "character progression", "meta-progression",
		"replayability", "turn-based roguelike", "real-time roguelike", "action roguelike",
		"strategy roguelike", "rpg roguelike", "shooter roguelike", "platformer roguelike",
		"card roguelike", "survival roguelike", "mystery dungeon", "traditional roguelike",
		"coffee break roguelike", "ascii roguelike", "tactical roguelike", "deck-building roguelike",
		"roguelike-metroidvania", "bullet hell roguelike",
	},
	"sandbox": {
		"sandbox", "open world", "crafting", "building", "voxel", "physics", "creative",
		"exploration", "survival", "procedural generation", "terraforming", "base-building",
		"resource management", "life simulation", "social simulation", "player-driven economy",
		"player-created content", "mod support", "multiplayer sandbox", "virtual world", "space sandbox",
		"historical sandbox", "fantasy sandbox", "sci-fi sandbox", "post-apocalyptic sandbox",
		"crime sandbox", "medieval sandbox", "western sandbox", "underwater sandbox", "playground",
		"simulation sandbox", "sandbox rpg",
	},
	"education": {
		"education", "educational", "learning", "science", "math", "language learning", "history",
		"geography", "programming", "typing", "quiz", "puzzle", "brain training", "memory", "logic",
		"problem-solving", "critical thinking", "creativity", "art", "music education",
		"physics simulation", "chemistry", "biology", "anatomy", "astronomy", "geology", "environmental",
		"social studies", "economics", "political science", "psychology", "philosophy", "literature",
		"grammar", "vocabulary", "foreign language", "sign language", "coding for kids",
	},
	"indie": {
		"indie", "experimental", "artistic", "minimalist", "pixel graphics", "hand-drawn", "stylized",
		"atmospheric", "surreal", "abstract", "quirky", "unique", "innovative", "niche", "cult classic",
		"short", "casual", "story-rich", "emotional", "thought-provoking", "philosophical", "political",
		"social commentary", "indie rpg", "indie platformer", "indie puzzle", "indie adventure",
		"indie horror", "indie strategy", "indie simulation", "indie roguelike", "indie multiplayer",
		"indie co-op", "indie sandbox",
	}}

type SteamRater struct {
	logger    *logger.AppLogger
	cfSvc     *cloudflare.CFService
	geminiSvc *gemini.GeminiService
}

func NewSteamRater(logger *logger.AppLogger) *SteamRater {
	genConfig := map[string]interface{}{
		"temperature":        0.1,
		"topP":               0.2,
		"topK":               1,
		"response_mime_type": "application/json",
	}

	cfSvc := cloudflare.NewCFService()
	geminiSvc := gemini.NewGeminiService(genConfig)

	return &SteamRater{
		logger:    logger,
		cfSvc:     cfSvc,
		geminiSvc: geminiSvc,
	}
}

func (s *SteamRater) GetSteamPageRating(spc SteamPageContent, se *gsheets.SheetsEntry) (*SteamPageRatingResult, error) {
	const scoreMult = 20
	var imgUrlContextList = s.ExtractImgUrlsGenerateText(&spc)

	spPromptContext := &SteamPagePromptCtx{
		Description:   spc.CapsuleDesc,
		AboutThisGame: spc.AboutGameText,
		Genres:        spc.Genres,
	}
	err := AddImgCaptionToCtx(spPromptContext, imgUrlContextList)
	if err != nil {
		s.logger.ErrorLog.Println(err.Error())
		return nil, err
	}

	finalPrompt := GetSteamPageEvalPrompt(spPromptContext)
	s.logger.InfoLog.Println("finished final prompt")
	s.logger.InfoLog.Println(finalPrompt)

	respBytes, err := s.geminiSvc.CallGeminiLLMApi(finalPrompt)
	if err != nil {
		s.logger.ErrorLog.Println(err.Error())
		return nil, err
	}
	rating := &LLMInnerResponse{}
	err = json.Unmarshal(respBytes, rating)
	if err != nil {
		s.logger.ErrorLog.Println(err.Error())
		return nil, err
	}
	s.logger.InfoLog.Println("finished generating gemini rating response")

	spscr := RateGameTags(spc.Genres, spc.Tags)

	descriptionScore, _ := strconv.ParseFloat(rating.Description.Score, 64)
	genresSectionScore, _ := strconv.ParseFloat(rating.Genres.Score, 64)
	tagsScore := int(spscr.Score[0])
	highlightImagesScore, _ := strconv.ParseFloat(rating.HighlightImageCaptions.Score, 64)
	aboutSectionScore, _ := strconv.ParseFloat(rating.AboutThisGame.Score, 64)
	capsuleImageScore, _ := strconv.ParseFloat(rating.CapsuleImageCaption.Score, 64)
	descriptionScore *= scoreMult
	genresSectionScore *= scoreMult
	tagsScoreF := float64(tagsScore) * scoreMult
	highlightImagesScore *= scoreMult
	aboutSectionScore *= scoreMult
	capsuleImageScore *= scoreMult

	// Define weights
	const (
		descriptionWeight     = 0.30
		genresWeight          = 0.10
		tagsWeight            = 0.10
		highlightImagesWeight = 0.10
		capsuleImageWeight    = 0.20
		aboutSectionWeight    = 0.20
	)

	// Calculate weighted scores
	weightedDescriptionScore := descriptionScore * descriptionWeight
	weightedGenresScore := genresSectionScore * genresWeight
	weightedTagsScore := tagsScoreF * tagsWeight
	weightedHighlightImagesScore := highlightImagesScore * highlightImagesWeight
	weightedAboutSectionScore := aboutSectionScore * aboutSectionWeight
	weightedCapsuleImageScore := capsuleImageScore * capsuleImageWeight

	totalWeightedScore := int(weightedDescriptionScore + weightedGenresScore + weightedTagsScore +
		weightedHighlightImagesScore + weightedAboutSectionScore + weightedCapsuleImageScore)

	// Set component names
	spscr.Component = "Tags"
	rating.Description.Component = "Description"
	rating.Genres.Component = "Genres"
	rating.HighlightImageCaptions.Component = "Highlight Images"
	rating.AboutThisGame.Component = "About Game"
	rating.CapsuleImageCaption.Component = "Capsule Image"

	// Update scores as strings
	rating.Description.Score = strconv.Itoa(int(descriptionScore))
	rating.Genres.Score = strconv.Itoa(int(genresSectionScore))
	spscr.Score = strconv.Itoa(int(tagsScoreF))
	rating.HighlightImageCaptions.Score = strconv.Itoa(int(highlightImagesScore))
	rating.AboutThisGame.Score = strconv.Itoa(int(aboutSectionScore))
	rating.CapsuleImageCaption.Score = strconv.Itoa(int(capsuleImageScore))

	// Create components slice
	steamPageComponentRatings := []SteamPageSingleComponentRating{
		rating.Description,
		*spscr,
		rating.HighlightImageCaptions,
		rating.Genres,
		rating.AboutThisGame,
		rating.CapsuleImageCaption,
	}

	// Create combined response
	steamPageRatingResult := &SteamPageRatingResult{
		FinalWeightedScore: totalWeightedScore,
		CapsuleUrl:         spc.CapsuleImgUrl,
		ComponentRatings:   steamPageComponentRatings,
	}

	ratingData, _ := json.Marshal(steamPageRatingResult)

	//assign needed sheets data
	se.Score = strconv.Itoa(totalWeightedScore)
	se.Rating = string(ratingData)
	se.Prompt = finalPrompt

	return steamPageRatingResult, nil
}

func RateGameTags(genres []string, tags []string) *SteamPageSingleComponentRating {
	var runningTotal int = 0
	var tagsNegFeedback []string
	var tagsPosFeedback []string

	var targetTags []string
	for _, genre := range genres {
		lGen := strings.ToLower(genre)
		if _, ok := genreToTags[lGen]; ok {
			lTags := genreToTags[lGen]
			targetTags = append(targetTags, lTags...)
		}
	}

	var foundTags []string
	for _, tag := range tags {
		lTag := strings.ToLower(tag)
		if _, ok := genreToTags[lTag]; ok {
			lTags := genreToTags[lTag]
			foundTags = append(foundTags, lTags...)
		}
	}

	if len(tags) >= 10 {
		runningTotal += 1
		tagsPosFeedback = append(tagsPosFeedback, "You have at least 10 tags which should increase search visibility.")
	} else {
		tagsNegFeedback = append(tagsNegFeedback, "Consider adding more than 10 tags. You should have anywhere from 15-25 tags for optical visibility results.")
	}

	if len(foundTags) >= 5 {
		runningTotal += 4
		tagsPosFeedback = append(tagsPosFeedback, "Your tags seem to align with your genre.")
	} else {
		runningTotal += 2
		tagsNegFeedback = append(tagsNegFeedback, "Consider adding more tags that align with your genre.")
	}

	spscr := &SteamPageSingleComponentRating{
		Score:              string(runningTotal),
		ActionableFeedback: strings.Join(tagsNegFeedback, " "),
		Strengths:          strings.Join(tagsPosFeedback, " "),
	}

	return spscr
}

func GetSteamPageEvalPrompt(ctx *SteamPagePromptCtx) string {
	genresString := strings.Join(ctx.Genres, ", ")
	highlightImageCaptionsString := strings.Join(ctx.HighlightImageCaptions, ",\n")

	exampleEvaluation := `Description context:
	Parse-O-Rhythm is a rhythm game about slashing errors in files to fix them. Slice and dice your way through files with nothing but the mouse and two buttons!
	Checklist:
	- Does it mention gameplay verbs?
	- Does it have a hook?
	- Does it mention at least one game genre?
	- Is it grammatically correct?
	Evaluation Results:
	{
		"description": {
			"score": "10",
			"actionablefeedback": "",
			"strengths": "The description effectively uses gameplay verbs such as 'slashing' and 'slice and dice,' includes a strong hook, mentions the rhythm game genre, and is grammatically correct. It concisely communicates the core gameplay while being engaging."
		}
	}`

	promptTemplate := `
		As a Steam page rating expert, you are tasked with evaluating a Steam page's content separated into components. Please follow the directions and rate the components on a scale of 1-5 based solely on the checklist criteria below.

        1. Use the following scoring system:
            - 5 points: All checklist criteria are met for the component.
            - 4 points: Most checklist criteria are met for the component (3 out of 4, or 2 out of 3 for 3-item lists).
            - 3 points: About half of the checklist criteria are met (approximately 50%% or 2 out of 4).
            - 2 points: Some checklist criteria are met (approximately 25%% or 1 out of 4).
            - 1 point: Very few checklist criteria are met.

        2. Evaluate each of the components below based on each individual context:
            Description context:
            %s
			Checklist:
			- Does it mention gameplay verbs?
			- Does it have a hook?
			- Does it mention at least one game genre?
			- Is it grammatically correct?
			Here is an example evaluation:
			%s

			AboutThisGame context:
			%s
			Checklist:
			- Does it mention key features and mechanics?
			- Does it explain what you do in the game and what the gameplay is like?
			- Does it contain a call to action regarding directing players to engage with the game?
			- Does it briefly explain the game's core concept or unique selling point?

			Genres context:
			Genres: %s
			Checklist:
			- Do the listed genres align with the game's Description component?
			- Do the listed genres align with the game's AboutThisGame component?
			- Do the listed genres mention any of the following genres: %s

			HighlightImage context (image to text descriptions, so be flexible and don't grade it harshly):
			%s
			Checklist:
			- Are the images context described well?
			- Are the descriptions concise and straight to the point?
			- Are there elements in the context that would intrigue potential players?
			- Does the context hint at the game's core mechanics or unique features?
			- Do the images context collectively showcase various aspects of the game (e.g., environment, characters, gameplay)?

			CapsuleImage context:
			%s
			Checklist:
			- Does it have the game title in the context text?
			- Does it show a theme or atmosphere in the background?

		3. Please provide your evaluation in the following JSON format for the output:

			json
			{
				"description": {
					"score": "",
					"actionablefeedback": "",
					"strengths": ""
				},
				"aboutThisGame": {
					"score": "",
					"actionablefeedback": "",
					"strengths": ""
				},
				"genres": {
					"score": "",
					"actionablefeedback": "",
					"strengths": ""
				},
				"highlightImageCaptions": {
					"score": "",
					"actionablefeedback": "",
					"strengths": ""
				},
				"capsuleImageCaption": {
					"score": "",
					"actionablefeedback": "",
					"strengths": ""
				}
			}
		4. Remember to adhere to the rules below:
			- The score should be based solely on the checklist criteria.
			- For components with 3 or 4 checklist items, a score of 4 is awarded if 2 or 3 criteria are met.
			- Provide actionable feedback for any unmet criteria.
			- Sentences should be at least 60 characters long and include specific suggestions for improvement.
		`

	return fmt.Sprintf(promptTemplate,
		ctx.Description,
		exampleEvaluation,
		ctx.AboutThisGame,
		genresString,
		genresString,
		highlightImageCaptionsString,
		ctx.CapsuleImageCaption,
	)
}

func AddImgCaptionToCtx(sppc *SteamPagePromptCtx, spiList []SteamPageImg) error {
	if len(spiList) == 0 {
		return fmt.Errorf("no images found")
	}

	for i := range spiList {
		if spiList[i].ImgType == "capsule" {
			sppc.CapsuleImageCaption = spiList[i].ImgCaption
		} else {
			sppc.HighlightImageCaptions = append(sppc.HighlightImageCaptions, spiList[i].ImgCaption)
		}
	}

	return nil
}

func (s *SteamRater) ExtractImgUrlsGenerateText(spc *SteamPageContent) []SteamPageImg {
	var imgUrlContextList []SteamPageImg
	for _, imgUrl := range spc.HighlightImgUrls[:3] {
		img := SteamPageImg{
			Url:     imgUrl,
			ImgType: "highlight",
		}
		imgUrlContextList = append(imgUrlContextList, img)
	}
	capImg := SteamPageImg{
		Url:     spc.CapsuleImgUrl,
		ImgType: "capsule",
	}
	imgUrlContextList = append(imgUrlContextList, capImg)

	//If we need to limit concurrent downloads, we can use a channel
	var wg sync.WaitGroup
	for i := range imgUrlContextList {
		wg.Add(1)

		go func(spi *SteamPageImg) {
			defer wg.Done()
			s.logger.InfoLog.Println("downloading img from url:", spi.Url)
			DownloadSteamImg(spi)
		}(&imgUrlContextList[i])
	}
	wg.Wait()
	s.logger.InfoLog.Println("successful extraction and generation of img text")

	//creating slice to pass underlying array reference
	imgUrlSlice := imgUrlContextList[:]
	s.ProcessImgCaptions(imgUrlSlice, spc)

	return imgUrlContextList
}

func (s *SteamRater) ProcessImgCaptions(imgUrlContextList []SteamPageImg, spc *SteamPageContent) {
	var wg sync.WaitGroup
	for i := range imgUrlContextList {
		wg.Add(1)

		go func(spi *SteamPageImg) {
			defer wg.Done()
			var imgContext string

			if spi.ImgType == "highlight" {
				imgContext = fmt.Sprintf("Describe this video game screenshot in THREE sentences (genres: %s), focusing on key gameplay elements, characters, environment, and any unique features that stand out.", strings.Join(spc.Genres, ", "))
			} else {
				imgContext = "This is a video game steam page capsule image, What's the title? What's the theme of the background like? Describe it in two short and concise sentences."
			}

			err := s.ProcessImgToText(spi, imgContext)
			if err != nil {
				s.logger.ErrorLog.Println(err.Error())
			}
		}(&imgUrlContextList[i])

	}
	wg.Wait()
	s.logger.InfoLog.Println("finished processing img captions")
}

func (s *SteamRater) ProcessImgToText(spi *SteamPageImg, imgContext string) error {
	type ImgToTextResponse struct {
		Result   Description `json:"result"`
		Success  bool        `json:"success"`
		Errors   []string    `json:"errors,omitempty"`   // omitempty for cleaner JSON if no errors
		Messages []string    `json:"messages,omitempty"` // omitempty is optional here as well
	}

	imgInts := make([]int, len(spi.ImgBytes))
	for i, b := range spi.ImgBytes {
		imgInts[i] = int(b)
	}

	inputData := ImgToTextPayload{
		Image:     imgInts,
		Prompt:    imgContext,
		MaxTokens: 256,
	}

	jsonInput, err := json.Marshal(inputData)
	if err != nil {
		s.logger.ErrorLog.Println(err.Error())
		return err
	}

	bodyBytes, err := s.cfSvc.CallImgToTextApi(jsonInput)
	if err != nil {
		s.logger.ErrorLog.Println(err.Error())
		return err
	}

	imgResponse := ImgToTextResponse{}
	err = json.Unmarshal(bodyBytes, &imgResponse)
	if err != nil {
		log.Println("Error unmarshaling JSON:", err)
		return err
	}

	log.Printf("Img description %s\n", imgResponse.Result.Description)
	spi.ImgCaption = imgResponse.Result.Description
	return nil
}

func DownloadSteamImg(spi *SteamPageImg) {
	resp, err := http.Get(spi.Url)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Println("Could not download img")
	}

	imgBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Could not read downloaded img bytes")
	}
	spi.ImgBytes = imgBytes
}

type SteamPagePromptCtx struct {
	Description            string   `json:"description"`
	AboutThisGame          string   `json:"aboutThisGame"`
	Genres                 []string `json:"genres"`
	HighlightImageCaptions []string `json:"highlightImageCaptions"`
	CapsuleImageCaption    string   `json:"capsuleImageCaption"`
}

type SteamPageRatingResult struct {
	FinalWeightedScore int                              `json:"finalWeightedScore"`
	CapsuleUrl         string                           `json:"capsuleUrl"`
	ComponentRatings   []SteamPageSingleComponentRating `json:"componentRatings"`
}

type LLMInnerResponse struct {
	Description            SteamPageSingleComponentRating `json:"description"`
	AboutThisGame          SteamPageSingleComponentRating `json:"aboutThisGame"`
	Genres                 SteamPageSingleComponentRating `json:"genres"`
	HighlightImageCaptions SteamPageSingleComponentRating `json:"highlightImageCaptions"`
	CapsuleImageCaption    SteamPageSingleComponentRating `json:"capsuleImageCaption"`
}

type SteamPageSingleComponentRating struct {
	Component          string `json:"component,omitempty"`
	Score              string `json:"score"`
	ActionableFeedback string `json:"actionablefeedback"`
	Strengths          string `json:"strengths,omitempty"`
}

type SteamPageImg struct {
	Url        string
	ImgType    string
	ImgBytes   []byte
	ImgCaption string
}

type ImgToTextPayload struct {
	Image     []int  `json:"image"`
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens"`
}

type Description struct {
	Description string `json:"description"`
}
