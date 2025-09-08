
CLASSES = {
    [0] = "ffabd473", --Hunter (Охотник)-
    [1] = "ff8788ee", --Warlock (Чернокнижник)-
    [2] = "ffffffff", --Priest (Жрец)-
    [3] = "fff58cba", --Paladin (Паладин)-
    [4] = "ff3fc7eb", --Mage (Маг)-
    [5] = "fffff569", --Rogue (Разбойник)-
    [6] = "ffff7d0a", --Druid (Друид)-
    [7] = "ff0070de", --Shaman (Шаман)-
    [8] = "ffc79c6e", --Warrior (Воин)-
    [9] = "ffc41f3b", --Death knight (Рыцарь смерти)-
    [10] = "4ec0fffe", --captions
    [11] = "ffc41f3b", --horde
    [12] = "ff3fc7eb", --alliance
    [13] = "00ff96ff"  --guild
}

RACES = {
    [0] = "Человек",
    [1] = "Дворф",
    [2] = "Ночной эльф",
    [3] = "Гном",
    [4] = "Дреней",
    [5] = "Орк",
    [6] = "Нежить",
    [7] = "Таурен",
    [8] = "Тролль",
    [9] = "Кровавый эльф"
}

local function TEXT_COLOR(TEXT, INDEX)
    return ("|c%s%s|r"):format(CLASSES[INDEX], TEXT)
end

local function TEXT_RACE(CHARACTER)
    if CHARACTER[4]>=5 then
        return TEXT_COLOR(RACES[CHARACTER[4]], 11)
    else
        return TEXT_COLOR(RACES[CHARACTER[4]], 12)
    end
end

local function TEXT_GUILD(CHARACTER)
    if CHARACTER[5] == "" then
        return ""
    else
        return TEXT_COLOR("<"..CHARACTER[5]..">", 10)
    end
end

local function TEXT_NAME(CHARACTER)
    return ("|Hplayer:%s|h%s|h"):format(CHARACTER[1], TEXT_COLOR(CHARACTER[1], CHARACTER[6]))
end

local function TEXT_LEVEL(CHARACTER)
    return TEXT_COLOR("["..CHARACTER[2].."]", 12)
end

local function TEXT_GS(CHARACTER)
    return TEXT_COLOR(CHARACTER[3].." GS", 10)
end

local function TEXT_CHARACTER(CHARACTER)
    print(TEXT_LEVEL(CHARACTER).." "..TEXT_NAME(CHARACTER).." "..TEXT_GS(CHARACTER).." "..TEXT_RACE(CHARACTER).." "..TEXT_GUILD(CHARACTER))
end

local function PRINT_ARRAY(CHARACTERS, ACCOUNT)
    local COUNTER = 0
    print(" ")
    print(TEXT_COLOR("Все персонажи аккаунта "..TEXT_COLOR(ACCOUNT, 3), 10))
    print(" ")
    for index, CHARACTER in ipairs(CHARACTERS) do
        TEXT_CHARACTER(CHARACTER)
        COUNTER = COUNTER +1
    end
    print(" ")
    print(TEXT_COLOR("Найдено персонажей: "..TEXT_COLOR(COUNTER, 3), 10))
    print(" ")
end

SLASH_ARMORY1, SLASH_ARMORY2 = '/ar', '/armory'

SlashCmdList["ARMORY"] = function (MESSAGE)
    print(" ")
    print(TEXT_COLOR("Isengard Armory", 10))
    print(TEXT_COLOR("База от <%- date %>", 10))
    print(TEXT_COLOR("Cодержит аккаунтов - "..TEXT_COLOR("<%- accounts %>", 3)..", персонажей - "..TEXT_COLOR("<%- characters %>", 3), 10))

    local CURRENT_ACCOUNT, CURRENT_CHARACTERS, ACCOUNT, CHARACTERS = nil

    local FIND = nil

    local TARGET, REALM = UnitName("target")

    if (MESSAGE ~= "") then
        FIND = MESSAGE
    elseif (TARGET ~= nil) then
        FIND = TARGET
    else
        print(TEXT_COLOR("Параметры поиска не установлены.", 9))
        return 0
    end

    DATABASES={<%
        let string = "";
        for (let i = 0; i < index; i++) {
            string += `DATABASE${i}, `;
        }
        print(string.slice(0, string.length - 2));
    %>}
    for f = 1, <% print(index) %> do
        for key, value in pairs(DATABASES[f]) do
            CURRENT_ACCOUNT = key
            CURRENT_CHARACTERS = DATABASES[f][CURRENT_ACCOUNT]
            for key, value in pairs(CURRENT_CHARACTERS) do
                LOWER = strlower(FIND)
                if (strlower(value[1]) == LOWER) then
                    ACCOUNT = CURRENT_ACCOUNT
                    CHARACTERS = CURRENT_CHARACTERS
                    break
                end
            end
        end
    end

    if ACCOUNT == nil then
        print(TEXT_COLOR("По запросу ", 9)..TEXT_COLOR(FIND, 10)..TEXT_COLOR(" ничего не найдено.", 9))
    else
        local SORTED = {}
        for key, value in pairs(CHARACTERS) do
            table.insert(SORTED, value)
        end
        sort(SORTED, function(a, b) return a[3]>b[3] end)
        PRINT_ARRAY(SORTED, ACCOUNT)
    end
end