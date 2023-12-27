package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type pokedex struct {
	pokemon map[string]pokemonInformation
}

type cliCommand struct {
	name        string
	description string
	callback    func(cfg *config, areaToExplore, pokemonName string, userPokedex *pokedex) error
}

type config struct {
	Previous *string
	Next     string
	cache    Cache
}

type LocationNamedArea struct {
	EncounterMethodRates []struct {
		EncounterMethod struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"encounter_method"`
		VersionDetails []struct {
			Rate    int `json:"rate"`
			Version struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"encounter_method_rates"`
	GameIndex int `json:"game_index"`
	ID        int `json:"id"`
	Location  struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"location"`
	Name  string `json:"name"`
	Names []struct {
		Language struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"language"`
		Name string `json:"name"`
	} `json:"names"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
		VersionDetails []struct {
			EncounterDetails []struct {
				Chance          int   `json:"chance"`
				ConditionValues []any `json:"condition_values"`
				MaxLevel        int   `json:"max_level"`
				Method          struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				} `json:"method"`
				MinLevel int `json:"min_level"`
			} `json:"encounter_details"`
			MaxChance int `json:"max_chance"`
			Version   struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"pokemon_encounters"`
}

type pokemonLocationArea struct {
	Count    int     `json:"count"`
	Next     string  `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type pokemonInformation struct {
	Abilities []struct {
		Ability struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"ability"`
		IsHidden bool `json:"is_hidden"`
		Slot     int  `json:"slot"`
	} `json:"abilities"`
	BaseExperience int `json:"base_experience"`
	Forms          []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"forms"`
	GameIndices []struct {
		GameIndex int `json:"game_index"`
		Version   struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"version"`
	} `json:"game_indices"`
	Height    int `json:"height"`
	HeldItems []struct {
		Item struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"item"`
		VersionDetails []struct {
			Rarity  int `json:"rarity"`
			Version struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"held_items"`
	ID                     int    `json:"id"`
	IsDefault              bool   `json:"is_default"`
	LocationAreaEncounters string `json:"location_area_encounters"`
	Moves                  []struct {
		Move struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"move"`
		VersionGroupDetails []struct {
			LevelLearnedAt  int `json:"level_learned_at"`
			MoveLearnMethod struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"move_learn_method"`
			VersionGroup struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version_group"`
		} `json:"version_group_details"`
	} `json:"moves"`
	Name          string `json:"name"`
	Order         int    `json:"order"`
	PastAbilities []any  `json:"past_abilities"`
	PastTypes     []any  `json:"past_types"`
	Species       struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"species"`
	Sprites struct {
		BackDefault      string `json:"back_default"`
		BackFemale       any    `json:"back_female"`
		BackShiny        string `json:"back_shiny"`
		BackShinyFemale  any    `json:"back_shiny_female"`
		FrontDefault     string `json:"front_default"`
		FrontFemale      any    `json:"front_female"`
		FrontShiny       string `json:"front_shiny"`
		FrontShinyFemale any    `json:"front_shiny_female"`
		Other            struct {
			DreamWorld struct {
				FrontDefault string `json:"front_default"`
				FrontFemale  any    `json:"front_female"`
			} `json:"dream_world"`
			Home struct {
				FrontDefault     string `json:"front_default"`
				FrontFemale      any    `json:"front_female"`
				FrontShiny       string `json:"front_shiny"`
				FrontShinyFemale any    `json:"front_shiny_female"`
			} `json:"home"`
			OfficialArtwork struct {
				FrontDefault string `json:"front_default"`
				FrontShiny   string `json:"front_shiny"`
			} `json:"official-artwork"`
			Showdown struct {
				BackDefault      string `json:"back_default"`
				BackFemale       any    `json:"back_female"`
				BackShiny        string `json:"back_shiny"`
				BackShinyFemale  any    `json:"back_shiny_female"`
				FrontDefault     string `json:"front_default"`
				FrontFemale      any    `json:"front_female"`
				FrontShiny       string `json:"front_shiny"`
				FrontShinyFemale any    `json:"front_shiny_female"`
			} `json:"showdown"`
		} `json:"other"`
		Versions struct {
			GenerationI struct {
				RedBlue struct {
					BackDefault      string `json:"back_default"`
					BackGray         string `json:"back_gray"`
					BackTransparent  string `json:"back_transparent"`
					FrontDefault     string `json:"front_default"`
					FrontGray        string `json:"front_gray"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"red-blue"`
				Yellow struct {
					BackDefault      string `json:"back_default"`
					BackGray         string `json:"back_gray"`
					BackTransparent  string `json:"back_transparent"`
					FrontDefault     string `json:"front_default"`
					FrontGray        string `json:"front_gray"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"yellow"`
			} `json:"generation-i"`
			GenerationIi struct {
				Crystal struct {
					BackDefault           string `json:"back_default"`
					BackShiny             string `json:"back_shiny"`
					BackShinyTransparent  string `json:"back_shiny_transparent"`
					BackTransparent       string `json:"back_transparent"`
					FrontDefault          string `json:"front_default"`
					FrontShiny            string `json:"front_shiny"`
					FrontShinyTransparent string `json:"front_shiny_transparent"`
					FrontTransparent      string `json:"front_transparent"`
				} `json:"crystal"`
				Gold struct {
					BackDefault      string `json:"back_default"`
					BackShiny        string `json:"back_shiny"`
					FrontDefault     string `json:"front_default"`
					FrontShiny       string `json:"front_shiny"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"gold"`
				Silver struct {
					BackDefault      string `json:"back_default"`
					BackShiny        string `json:"back_shiny"`
					FrontDefault     string `json:"front_default"`
					FrontShiny       string `json:"front_shiny"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"silver"`
			} `json:"generation-ii"`
			GenerationIii struct {
				Emerald struct {
					FrontDefault string `json:"front_default"`
					FrontShiny   string `json:"front_shiny"`
				} `json:"emerald"`
				FireredLeafgreen struct {
					BackDefault  string `json:"back_default"`
					BackShiny    string `json:"back_shiny"`
					FrontDefault string `json:"front_default"`
					FrontShiny   string `json:"front_shiny"`
				} `json:"firered-leafgreen"`
				RubySapphire struct {
					BackDefault  string `json:"back_default"`
					BackShiny    string `json:"back_shiny"`
					FrontDefault string `json:"front_default"`
					FrontShiny   string `json:"front_shiny"`
				} `json:"ruby-sapphire"`
			} `json:"generation-iii"`
			GenerationIv struct {
				DiamondPearl struct {
					BackDefault      string `json:"back_default"`
					BackFemale       any    `json:"back_female"`
					BackShiny        string `json:"back_shiny"`
					BackShinyFemale  any    `json:"back_shiny_female"`
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"diamond-pearl"`
				HeartgoldSoulsilver struct {
					BackDefault      string `json:"back_default"`
					BackFemale       any    `json:"back_female"`
					BackShiny        string `json:"back_shiny"`
					BackShinyFemale  any    `json:"back_shiny_female"`
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"heartgold-soulsilver"`
				Platinum struct {
					BackDefault      string `json:"back_default"`
					BackFemale       any    `json:"back_female"`
					BackShiny        string `json:"back_shiny"`
					BackShinyFemale  any    `json:"back_shiny_female"`
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"platinum"`
			} `json:"generation-iv"`
			GenerationV struct {
				BlackWhite struct {
					Animated struct {
						BackDefault      string `json:"back_default"`
						BackFemale       any    `json:"back_female"`
						BackShiny        string `json:"back_shiny"`
						BackShinyFemale  any    `json:"back_shiny_female"`
						FrontDefault     string `json:"front_default"`
						FrontFemale      any    `json:"front_female"`
						FrontShiny       string `json:"front_shiny"`
						FrontShinyFemale any    `json:"front_shiny_female"`
					} `json:"animated"`
					BackDefault      string `json:"back_default"`
					BackFemale       any    `json:"back_female"`
					BackShiny        string `json:"back_shiny"`
					BackShinyFemale  any    `json:"back_shiny_female"`
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"black-white"`
			} `json:"generation-v"`
			GenerationVi struct {
				OmegarubyAlphasapphire struct {
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"omegaruby-alphasapphire"`
				XY struct {
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"x-y"`
			} `json:"generation-vi"`
			GenerationVii struct {
				Icons struct {
					FrontDefault string `json:"front_default"`
					FrontFemale  any    `json:"front_female"`
				} `json:"icons"`
				UltraSunUltraMoon struct {
					FrontDefault     string `json:"front_default"`
					FrontFemale      any    `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale any    `json:"front_shiny_female"`
				} `json:"ultra-sun-ultra-moon"`
			} `json:"generation-vii"`
			GenerationViii struct {
				Icons struct {
					FrontDefault string `json:"front_default"`
					FrontFemale  any    `json:"front_female"`
				} `json:"icons"`
			} `json:"generation-viii"`
		} `json:"versions"`
	} `json:"sprites"`
	Stats []struct {
		BaseStat int `json:"base_stat"`
		Effort   int `json:"effort"`
		Stat     struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Slot int `json:"slot"`
		Type struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"type"`
	} `json:"types"`
	Weight int `json:"weight"`
}

type Cache struct {
	cache    map[string]cacheEntry
	mu       sync.Mutex
	interval time.Duration
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

func NewCache(interval time.Duration) *Cache {
	newCache := &Cache{
		cache:    make(map[string]cacheEntry),
		interval: interval,
	}
	go newCache.reapLoop()
	return newCache
}

func (upok *pokedex) Add(pokemonName string, pokemonInfo pokemonInformation) {
	upok.pokemon[pokemonName] = pokemonInfo
}

func (upok *pokedex) Get(pokemonName string) (pokemonInformation, bool) {
	pokName, ok := upok.pokemon[pokemonName]
	if !ok {
		return pokemonInformation{}, false
	}
	return pokName, true
}

func (cache *Cache) Add(key string, val []byte) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.cache[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}

func (cache *Cache) Get(key string) ([]byte, bool) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	val, ok := cache.cache[key]
	if !ok {
		return nil, false
	}
	return val.val, true
}

func (cache *Cache) reapLoop() {
	ticker := time.NewTicker(cache.interval)
	defer ticker.Stop()
	for tick := range ticker.C {
		cache.mu.Lock()
		for k, v := range cache.cache {
			if tick.Sub(v.createdAt) > cache.interval {
				delete(cache.cache, k)
			}
		}
		cache.mu.Unlock()
	}
}

func cmdHelp(cfg *config, areaToExplore, pokemonName string, userPokedex *pokedex) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()
	for _, cmd := range getCommands() {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}
	return nil
}

func cmdExit(cfg *config, areaToExplore, pokemonName string, userPokedex *pokedex) error {
	os.Exit(0)
	return nil
}

// TODO: Look for a better way to avoid repeating the code
func cmdMap(cfg *config, areaToExplore, pokemonName string, userPokedex *pokedex) error {
	nextURL := cfg.Next
	body := getData(nextURL, &cfg.cache)
	locationResponse := pokemonLocationArea{}
	errUnmarshall := json.Unmarshal(body, &locationResponse)
	if errUnmarshall != nil {
		return errors.New("There has been an issue unmarshall ")
	}
	for _, loc := range locationResponse.Results {
		fmt.Println(loc.Name)
	}
	cfg.Previous, cfg.Next = locationResponse.Previous, locationResponse.Next
	return nil
}

func cmdMapb(cfg *config, areaToExplore, pokemonName string, userPokedex *pokedex) error {
	previousURL := cfg.Previous
	if previousURL == nil {
		return errors.New("We are already on the first page")
	}
	body := getData(*previousURL, &cfg.cache)
	locationResponse := pokemonLocationArea{}
	errUnmarshall := json.Unmarshal(body, &locationResponse)
	if errUnmarshall != nil {
		return errors.New("There has been an issue unmarshall ")
	}
	for _, loc := range locationResponse.Results {
		fmt.Println(loc.Name)
	}
	cfg.Previous, cfg.Next = locationResponse.Previous, locationResponse.Next
	return nil
}

func cmdExplore(cfg *config, areaToExplore, pokemonName string, userPokedex *pokedex) error {
	urlSplit := strings.Split(cfg.Next, "?")
	url := urlSplit[0] + areaToExplore + "?" + urlSplit[1]
	if len(areaToExplore) < 3 {

		return errors.New("Please, insert a valid area name.")
	}
	fmt.Println("Exploring", areaToExplore, "...")
	body := getData(url, &cfg.cache)
	locationNamedArea := LocationNamedArea{}
	errUnmarshall := json.Unmarshal(body, &locationNamedArea)
	if errUnmarshall != nil {
		errorAreaMsg := fmt.Sprintf("Could not get information about %s area...", areaToExplore)
		return errors.New(errorAreaMsg)
	}
	if len(locationNamedArea.PokemonEncounters) < 1 {
		return errors.New("No Pokemon found")
	}
	fmt.Println("Found Pokemon:")
	for _, pok := range locationNamedArea.PokemonEncounters {
		fmt.Println("- ", pok.Pokemon.Name)
	}
	return nil
}

func cmdCatch(cfg *config, areaToExplore, pokemonName string, userPokedex *pokedex) error {
	pokemonInfoURL := "https://pokeapi.co/api/v2/pokemon/" + pokemonName
	body := getData(pokemonInfoURL, &cfg.cache)
	pokemonInformation := pokemonInformation{}
	errUnmarshall := json.Unmarshal(body, &pokemonInformation)
	if errUnmarshall != nil {
		errorAreaMsg := fmt.Sprintf("Could not get information about %s area...", pokemonName)
		return errors.New(errorAreaMsg)
	}
	randCatchProb := rand.Intn(100)
	fmt.Println("Throwing a Pokeball at pikachu...")
	if pokemonInformation.BaseExperience > randCatchProb {
		_, ok := userPokedex.Get(pokemonName)
		if !ok {
			userPokedex.Add(pokemonName, pokemonInformation)
			fmt.Println(pokemonName, "was caught!")
		} else {
			fmt.Println("You already have this Pokemon...")
		}
	} else {
		fmt.Println(pokemonName, "escaped!")
	}
	return nil
}

func cmdInspect(cfg *config, areaToExplore, pokemonName string, userPokedex *pokedex) error {
	pok, ok := userPokedex.Get(pokemonName)
	if !ok {
		return errors.New("you have not caught that pokemon")
	} else {
		fmt.Println("Name:", pok.Name)
		fmt.Println("Height:", pok.Height)
		fmt.Println("Weight:", pok.Weight)
		fmt.Println("Stats:")
		for _, stat := range pok.Stats {
			fmt.Println("\t -", stat.Stat.Name, ":", stat.BaseStat)
		}
		fmt.Println("Types:")
		for _, typ := range pok.Types {
			fmt.Println("\t -", typ.Type.Name)
		}
	}
	return nil
}

func cmdPokedex(cfg *config, areaToExplore, pokemonName string, userPokedex *pokedex) error {
	if len(userPokedex.pokemon) < 1 {
		fmt.Println("You have not caught a Pokemon yet.")
	}
	fmt.Println("Your Pokedex:")
	for _, pok := range userPokedex.pokemon {
		fmt.Println("\t -", pok.Name)
	}
	return nil
}

func getData(url string, cache *Cache) []byte {
	var body []byte
	body, ok := cache.Get(url)
	if !ok {
		body = makeRequest(url)
		cache.Add(url, body)
	}
	return body
}

func makeRequest(url string) []byte {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	body, err := io.ReadAll(res.Body)
	errClose := res.Body.Close()
	if res.StatusCode > 299 || errClose != nil {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		log.Fatal(err)
	}
	return body
}

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    cmdHelp,
		},
		"map": {
			name:        "map",
			description: "Display the name of 20 locations in the Pokemon world",
			callback:    cmdMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Display the name of the previous 20 locations in the Pokemon world",
			callback:    cmdMapb,
		},
		"explore": {
			name:        "explore",
			description: "Get a list of all the Pokémon in a given area.",
			callback:    cmdExplore,
		},
		"catch": {
			name:        "catch",
			description: "Catch a Pokemon by name.",
			callback:    cmdCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "See details about a Pokemon if it has been captured",
			callback:    cmdInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "See the list of all caught Pokemon",
			callback:    cmdPokedex,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    cmdExit,
		},
	}
}

func main() {
	pageTracker := config{
		Next:     "https://pokeapi.co/api/v2/location-area/?offset=0&limit=20",
		Previous: nil,
		cache:    *NewCache(100 * time.Second),
	}
	userPokedex := &pokedex{pokemon: make(map[string]pokemonInformation)}
	for {
		fmt.Print("Pokedex > ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		if len(scanner.Text()) == 0 {
			continue
		}
		cmdExp := strings.Split(scanner.Text(), " ")
		cmd, ok := getCommands()[cmdExp[0]]
		if ok {
			areaToExplore := ""
			pokemonName := ""
			if cmdExp[0] == "explore" && len(cmdExp) > 1 {
				areaToExplore = cmdExp[1]
			}
			if cmdExp[0] == "catch" && len(cmdExp) > 1 || cmdExp[0] == "inspect" {
				pokemonName = cmdExp[1]
			}
			err := cmd.callback(&pageTracker, areaToExplore, pokemonName, userPokedex)
			if err != nil {
				fmt.Println(err)
			}
			continue
		} else {
			fmt.Println("Command not found")
			continue
		}
	}
}
