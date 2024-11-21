package gamedocgen

import (
	"encoding/json"
	"fmt"
	"gdrsapi/external/gemini"
	"gdrsapi/pkg/logger"
)

type GameDesignDocGen struct {
	geminiSvc *gemini.GeminiService
	logger    *logger.AppLogger
}

func NewgdDocGen(logger *logger.AppLogger) *GameDesignDocGen {
	simpleCfg := map[string]interface{}{
		"temperature":        0.8,
		"response_mime_type": "application/json",
	}
	geminiSvc := gemini.NewGeminiService(simpleCfg)

	return &GameDesignDocGen{
		logger:    logger,
		geminiSvc: geminiSvc,
	}
}

type BasicGameDesignDocContent struct {
	Overview     string   `json:"overview"`
	CoreGameplay []string `json:"coreGameplay"`
	KeyFeatures  []string `json:"keyFeatures"`
	ArtStyle     []string `json:"artStyle"`
}

type StarterGameDesignDocContent struct {
	Description           string   `json:"description"`
	UniqueSellingPoint    []string `json:"uniqueSellingPoint"`
	GameplayLoop          []string `json:"gameplayLoop"`
	CoreMechanics         []string `json:"coreMechanics"`
	ObjectiveAndEndgoal   []string `json:"objectiveAndEndgoal"`
	ArtStyleAndAtmosphere []string `json:"artStyleAndAtmosphere"`
}

func (g *GameDesignDocGen) GenerateGameDesignDoc(gameTitle string, gameDescription string, gameGenre string, template string) (interface{}, error) {
	prompt := GetGeneratePrompt(gameTitle, gameDescription, gameGenre, template)

	respBytes, err := g.geminiSvc.CallGeminiLLMApi(prompt)
	if err != nil {
		g.logger.ErrorLog.Println(err.Error())
		return nil, err
	}

	var doc interface{}
	switch template {
	case "basic":
		doc = &BasicGameDesignDocContent{}
	case "starter":
		doc = &StarterGameDesignDocContent{}
	}

	err = json.Unmarshal(respBytes, doc)
	if err != nil {
		g.logger.ErrorLog.Println(err.Error())
		return nil, err
	}
	return doc, nil
}

func (g *GameDesignDocGen) RegenerateGameDesignDoc(currentDocContent string, selection string, suggestion string, template string) (interface{}, error) {
	prompt := GetRegeneratePrompt(currentDocContent, selection, suggestion, template)

	respBytes, err := g.geminiSvc.CallGeminiLLMApi(prompt)
	if err != nil {
		g.logger.ErrorLog.Println(err.Error())
		return nil, err
	}

	var doc interface{}
	switch template {
	case "basic":
		doc = &BasicGameDesignDocContent{}
	case "starter":
		doc = &StarterGameDesignDocContent{}
	}

	g.logger.InfoLog.Printf("Gemini response: %s", string(respBytes))

	err = json.Unmarshal(respBytes, doc)
	if err != nil {
		g.logger.ErrorLog.Println(err.Error())
		return nil, err
	}
	return doc, nil
}

func GetGeneratePrompt(title, description, genre, template string) string {
	return fmt.Sprintf(`
	As a game design expert, you are tasked with creating a game design document just by being given a video game title, description/ideas, and genre. You are amazing at generating and writing game design documents. You have read thousands of books on game design and know all about game design gameplay, game mechanics, and unique features, so you will be extensive and creative with your work. Follow the instructions below.

	1. Use the video game ideas and context below:
		Here is the title of the game:
		%s

		Here is the description/ideas of the game:
		%s

		Here is the genre of the game:
		%s

	2. Generate a game design document that has the following components:
		%s

	3. Please provide the game design document content in the following JSON format for the output:
	`+"```json\n%s\n```"+`

	4. Here are examples of well-done game design documents:
		%s

	5. Remember to be creative, detailed, and consistent.`,
		title,
		description,
		genre,
		templateOutline[template],
		templateJsonFormats[template],
		templateExamples[template],
	)
}

func GetRegeneratePrompt(currentDocument, selection, suggestion, template string) string {
	return fmt.Sprintf(`
	As a game design expert, you are tasked with editing/refining a specific section of an existing game design document. You have extensive knowledge of game design, gameplay mechanics, and unique features. Use your expertise to do as you are asked on the selected section while maintaining consistency with the overall game concept.

	1. Review the current game design document content below:
	%s

	2. Focus on the following selected text to regenerate/update/delete. This is the text that needs to be updated (just this portion) or deleted depending on the suggestion in the third point right below:
	%s

	3. Consider this suggestion/demand or context for the regeneration/update/deletion. If it mentions to remove or delete the section, please do so:
	%s

	4. Update/edit the selected section, ensuring it:
	- Aligns with the overall game concept and style
	- Expands upon or improves the existing ideas
	- Incorporates the additional suggestion or context provided
	- Maintains a consistent tone and level of detail with the rest of the document

	5. Provide the regenerated content in the same format as the original section. If the regenerated content affects other sections of the document, briefly describe how those sections should be updated for consistency.

	6. Return the regenerated section in JSON format, including only the modified section(s) of the document. For example:
	`+"```json\n%s\n```"+`

	7. Remember to be creative, detailed, and consistent with the existing game design while incorporating improvements and suggestions.`,
		currentDocument,
		selection,
		suggestion,
		templateJsonFormats[template],
	)
}

func GetBasicExamples() string {
	example := `
	FIRST EXAMPLE:
	{
		"overview": "Survival Selection is a challenging strategy-survival game set in the dawn of human civilization. Players begin as a single individual with a specific skill set, tasked with building a thriving community in a harsh, unforgiving world. The game features permadeath mechanics, diverse biomes, and a unique progression system that emphasizes cooperation and specialization. As players expand their settlement and recruit new members, they must balance resource management, skill development, and exploration to ensure the survival and growth of their fledgling society.",
		"coreGameplay": [
			"Resource Gathering: Players collect food, water, and materials essential for survival",
			"Skill Development: Improve personal abilities and unlock new professions as players level up",
			"Community Building: Construct shelter, craft tools, and create a sustainable living environment",
			"NPC Recruitment: Attract and integrate new members with diverse skills into the community",
			"Exploration: Discover new biomes, resources, and technologies as the player's territory expands",
			"Survival Management: Balance individual and community needs while facing environmental challenges and conflicts",
			"Turn-Based System: Each action represents a day, requiring careful management of decisions to ensure survival"
		],
		"keyFeatures": [
			"Permadeath Mechanic: Characters have one life, and revival requires meeting specific, challenging requirements",
			"Dynamic Skill System: Players start with a preset skill and unlock additional professions based on skill progression",
			"NPC Recruitment: Recruit characters with unique skills, including rare and legendary NPCs with exceptional abilities",
			"Biome Progression: Unlock access to new environments through skill and technology development",
			"Community Synergy: Success relies on creating a balanced community where skills complement each other",
			"Tech Tree and Crafting: Unlock new technologies and craft items to overcome environmental challenges",
			"Random Events: Face unpredictable challenges like natural disasters, diseases, or hostile encounters",
			"Seasonal Cycles: Adapt to changing seasons that impact resource availability and survival strategies",
			"Exploration Missions: Send community members on expeditions to discover new resources and technologies",
			"Legacy System: When a community fails, certain bonuses or abilities carry over to the next playthrough"
		],
		"artStyle": [
			"Low-poly Environment: Colorful, low-poly world design with clear distinctions between biomes and resources",
			"Character Design: Simple but expressive characters with unique silhouettes that convey their skills and personalities",
			"Minimalist UI: Earthy tones and natural textures, inspired by ancient cave paintings and primitive art",
			"Dynamic Lighting: Creates atmospheric depth, with elements like sunsets and bioluminescent plants adding to the ambiance",
			"Stylized Animation: Emphasizes key actions while maintaining the low-poly aesthetic",
			"Weather Effects: Simplified but effective visuals for weather, from sandstorms to blizzards"
		]
	}
	SECOND EXAMPLE:
	{
		"overview": "Pocket Monster Land is a fast-paced RPG adventure game inspired by classic titles like Pokémon and Mario Party. Players embark on a journey across diverse islands, each with its own unique theme and challenges, to capture and train Pocket Monsters. The game features a dice-based movement system, similar to Monopoly or Mario Party, allowing for strategic decision-making and unexpected encounters. Players can explore hidden levels, uncover rare fossils, and battle gym leaders in a quest to become the ultimate Pocket Monster Master.",
		"coreGameplay": [
			"Dice Movement: Players roll a dice to determine their movement across the map, navigating through various environments and encountering Pocket Monsters",
			"Island Exploration: Each island has its own theme (e.g., lava, ocean, air) and features unique Pocket Monsters to capture",
			"Pocket Monster Capture: Players can capture Pocket Monsters by engaging in turn-based battles. Each Pocket Monster has its own strengths and weaknesses based on its type",
			"Training and Evolution: Players can train their Pocket Monsters to improve their stats and evolve them into stronger forms",
			"Gyms and Battles: Players must defeat gym leaders to earn badges and advance to the next island",
			"Hidden Levels and Secrets: Exploring the map can uncover hidden levels and secrets, such as rare fossils or bonus items",
			"Fossil Mini-Game: Players can participate in a mini-game to dig for fossils, which can be used to revive extinct Pocket Monsters",
			"Daily Island Rotation: A new random island is available each day, providing opportunities to encounter different Pocket Monsters and collect rare items",
			"Dungeon Crawler Levels: Certain areas feature dungeon crawler levels, where players must navigate through intricate mazes and solve puzzles to reach the exit"
		],
		"keyFeatures": [
			"Dice-based movement system: Adds an element of chance and strategy to the gameplay",
			"Diverse island themes: Provides variety and exploration opportunities",
			"Unique Pocket Monster capture mechanics: Offers a fresh take on traditional monster catching",
			"Fossil mini-game: Introduces a new layer of gameplay and collectible elements",
			"Daily island rotation: Ensures replayability and keeps the game fresh",
			"Fast-paced gameplay: Streamlines the experience and reduces grinding",
			"Hidden levels and secrets: Encourages exploration and discovery"
		],
		"artStyle": [
			"Colorful and vibrant: Emphasize the whimsical and adventurous nature of the game",
			"2D pixel art: Evokes a classic retro style while maintaining a modern aesthetic",
			"Detailed character designs: Capture the personality and charm of the Pocket Monsters",
			"Dynamic environments: Bring the diverse island themes to life"
		]
	}
	`
	return example
}

func GetStarterExamples() string {
	example := `
	FIRST EXAMPLE:
	{
		"description": "Survival Selection is a challenging strategy survival game set at the dawn of human civilization. Players navigate a grid-based world, starting as a lone individual with a specific skill set. The goal is to build a thriving community by recruiting diverse NPCs, each with unique abilities. As players expand their settlement and technology, they unlock new biomes to explore, facing increasingly difficult survival challenges. With permadeath mechanics and the need for strategic resource management, every decision is crucial in this unforgiving world where cooperation is key to survival and progress.",
		"uniqueSellingPoint": [
			"Permadeath mechanic with specific revival requirements, adding tension and strategic depth",
			"Dynamic NPC recruitment system with rare, legendary characters offering unique abilities",
			"Profession-based character progression system that encourages diversification and cooperation",
			"Biome unlocking system tied to technological and skill advancements, providing a sense of progression and exploration"
		],
		"gameplayLoop": [
			"Survive and gather resources in the starting biome",
			"Recruit NPCs to expand skill set and workforce",
			"Construct and upgrade buildings to improve settlement",
			"Research and craft new technologies",
			"Unlock and explore new biomes",
			"Face new challenges and gather rare resources",
			"Repeat steps 2-6, gradually expanding and strengthening the community"
		],
		"coreMechanics": [
			"Grid-based movement: Players move in a grid-based system, influencing resource gathering, combat, and building placement",
			"Resource Management: Efficiently gathering and managing resources is crucial for survival and expansion",
			"Crafting and Construction: Players craft tools, weapons, and build structures to improve their survival chances",
			"NPC Recruitment: Recruit NPCs from various backgrounds who bring unique skills and personalities to the settlement",
			"Skill Progression: Players level up skills and unlock new professions, allowing for specialization and greater efficiency",
			"Biome Exploration: Unlocking access to new biomes expands possibilities but introduces new challenges",
			"Survival and Death: The one-life system creates high tension and strategic decision-making. Revival requires specific actions or resource sacrifices",
			"Technology Progression: Unlocking technologies is critical for biome access and improving survival prospects"
],
		"objectiveAndEndgoal": [
			"Short-term: Establish a self-sustaining settlement in the starting biome",
			"Mid-term: Unlock and successfully colonize all available biomes",
			"Long-term: Achieve the highest level of technology and build a thriving, diverse community",
			"Ultimate goal: Survive for a set number of in-game years or reach a specific population milestone",
			"Optional challenges: Discover all legendary NPCs or unlock all possible technologies"
		],
		"artStyleAndAtmosphere": [
			"Vivid, slightly stylized 2D graphics with a top-down perspective",
			"Each biome has a distinct color palette and visual theme",
			"Character designs reflect primitive human aesthetics with clear profession-based visual cues",
			"Dynamic lighting system to represent time of day and weather conditions",
			"Ambient sound design featuring nature sounds specific to each biome",
			"Minimalistic UI with stone and wood textures to match the primitive setting",
			"Atmospheric music that evolves as the player progresses through different stages of civilization"
		]
	}
	SECOND EXAMPLE:
	{
		"description": "Dream Architect is a surreal puzzle-platformer where players take on the role of a Dream Walker - a being capable of manipulating the dreamscapes of sleeping individuals. Set in a world where dream disorders have become epidemic, players must navigate and reshape the abstract landscapes of others' dreams to cure their psychological ailments. Each level represents a different person's dreamscape, filled with manifestations of their anxieties, hopes, and memories that must be carefully rearranged and resolved through creative reality manipulation.",
		"uniqueSellingPoint": [
			"Reality-bending mechanics that allow players to rotate, reshape, and reimagine sections of the dream world",
			"Emotional resonance system where player actions create rippling effects throughout the dreamscape",
			"Dynamic dream logic that changes based on the dreamer's psychological state and memories",
			"Architectural puzzles that require both spatial reasoning and emotional intelligence to solve",
			"Every level tells the story of a different person's inner struggles and hopes"
		],
		"gameplayLoop": [
			"Enter a new patient's dream and analyze their dream environment",
			"Discover memory fragments scattered throughout the dreamscape",
			"Use reality-bending powers to manipulate the dream architecture",
			"Solve environmental puzzles that represent psychological barriers",
			"Balance the dreamer's emotional state through careful manipulation",
			"Connect fragmented memories to reveal the core issue",
			"Resolve the central dream conflict to cure the patient"
		],
		"coreMechanics": [
			"Dream Walking: Phase through different layers of the dream",
			"Reality Bending: Rotate, stretch, and transform dream environments",
			"Memory Echo: Replay and interact with captured memory fragments",
			"Emotional Resonance: Actions create ripple effects that influence dream stability",
			"Architecture Manipulation: Reshape dream structures to create new paths",
			"Time Dilation: Speed up or slow down sections of the dream",
			"Dream Logic: Use inconsistent physics and perspective shifts to solve puzzles",
			"Psychological Balance: Maintain harmony between different emotional aspects"
		],
		"objectiveAndEndgoal": [
			"Short-term: Successfully navigate and solve individual dream puzzles",
			"Mid-term: Cure patients by resolving their core dream conflicts",
			"Long-term: Uncover the source of the dream disorder epidemic",
			"Ultimate goal: Prevent the collapse of the collective dreamscape",
			"Optional challenges: Find hidden memories and alternate resolutions",
			"Meta-progression: Unlock new reality-bending abilities and dream-walking techniques"
		],
		"artStyleAndAtmosphere": [
			"Surrealist art style inspired by M.C. Escher and Salvador Dalí",
			"Floating geometry and impossible architecture that shifts and transforms",
			"Color palettes that morph based on emotional states and dream stability",
			"Particle effects that trace the flow of memories and emotions",
			"Dreamlike transitions between spaces using liquid geometry",
			"Abstract character designs that represent psychological archetypes",
			"Ambient soundscape that combines real-world sounds with surreal distortions",
			"Minimalist UI elements that appear as natural parts of the dream environment"
		]
	}
	`
	return example
}

var templateOutline = map[string]string{
	"basic": `
		- Overview
		- Core gameplay
		- Key features
		- Art style
		`,
	"starter": `
		- Description
		- Unique Selling Point
		- Gameplay Loop
		- Core Mechanics
		- Objective/End-goal
		- Art Style & Atmosphere
		`,
}

var templateJsonFormats = map[string]string{
	"basic": `{
		"overview": " ",
		"coreGameplay": "[]",
		"keyFeatures": "[]",
		"artStyle": "[]"
	}`,
	"starter": `{
		"description": " ",
		"uniqueSellingPoint": "[]",
		"gameplayLoop": "[]",
		"coreMechanics": "[]",
		"objectiveAndEndgoal": "[]",
		"artStyleAndAtmosphere": "[]"
	}`,
}

var templateExamples = map[string]string{
	"basic":   GetBasicExamples(),
	"starter": GetStarterExamples(),
}
