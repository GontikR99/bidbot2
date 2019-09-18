local eq=require "everquest"
local http=require "http"
local re=require "re"

-- Settings
local minimum_bid = 5
local alt_bid_cap = 100
local dkp_site="https://modusgelidus.gamerlaunch.com/rapid_raid/leaderboard.php"
local dkp_pattern="<a href=[^>]*>([^<]*)</a></td><td[^>]*><span class='dkp_current'>([0-9.,]*)</span>"

local solicitations={
    "Who wants %?",
    "You know you need %!",
    "I spy with my little eye, %.",
    "Baby needs a new %.",
    "You've never seen loot like this %.",
    "Dude!  Checkout %!",
    "The one!  The only! %!",
    "The myth!  The legend! %!",
    "Greetings raider!  You look like you could use a %.",
    "Just browsing?  Have you seen the % I just got in?",
    "% could be yours for a low, low price!",
    "Do I look fat in %?",
    "Anybody missing their % should send me a bid!",
    "How much D K P is % worth to you?",
    "Peanuts!  Popcorn!  %!",
    "New studies show, % causes violence.",
    "You show me yours, and I'll show you my %.",
    "It's a bird!  It's a plane!  No, it's %!",
    "Come and get % while it's hot!",
}

-- Utility: take a string and rewrite it to have the first letter capital, for display purposes
local function initial_cap(name)
    if name:len()==0 then
        return name
    end
    return name:sub(1,1):upper()..name:sub(2):lower()
end

-- Determine if the given character is an alt by looking at a guild dump.
local function isalt(charname, guilddump)
    record = guilddump[charname]
    return record~=nil and (
                    record.alt or
                    record.rank:lower()=="box/alt" or
                    record.rank:lower()=="officer box/alt")
end

-- Retrieve and parse the entire leaderboard to get everyone's DKP
local function getalldkp()
    local result={}
    local page = http.get(dkp_site)
    for name, dkptext in re.gmatch(page, dkp_pattern) do
        result[name:lower()]=tonumber(dkptext:gsub(",",""))
    end
    return result
end

-- Return a phrase trying to sell the item to the masses
function solicit(item)
    index=math.random(1, #solicitations)
    return solicitations[index]:gsub("%%", item)
end

-- Return the name of the character who should be charged when `charname` wins something.
-- While the spec expects only 1 argument, we define a function taking 2 arguments here. When called with one argument,
-- we will just get nil for the extra arguments, and will make calls to fill them
function getmain(charname, guilddump)
    if guilddump==nil then
        guilddump=eq.guildmembers()
    end
    if isalt(charname, guilddump) then
        return record.note:lower()
    else
        return charname
    end
end

-- Get the currently posted DKP for the specified character.
-- While the spec expects only 1 argument, we define a function taking 3 arguments here. When called with one argument,
-- we will just get nil for the extra arguments, and will make calls to fill them
function getdkp(charname, alldkp, guilddump)
    if alldkp==nil then
        alldkp=getalldkp()
    end
    mainname = getmain(charname, guilddump)
    if mainname~=nil then
        return alldkp[mainname]
    else
        return nil
    end
end

-- Determine if the `charname` is allowed to bid `quantity`, returning a description of the problem
-- if they're not allowed to bid.
function validatebid(charname, quantity)
    if quantity<minimum_bid then
        return "You must bid at least "..minimum_bid.." DKP."
    elseif quantity~=math.floor(quantity) then
        return "Please bid a whole number."
    else
        return nil
    end
end

-- Determine the winner(s) of an auction, and how to display the outcome.
-- Accepts
-- - bids: table mapping bidder (string) to bid (number)
-- - count: number of items available
-- Must return:
-- - price (number)
-- - winners (list of strings)
-- - ordered list of pairs with [1]=description of bidder, [2] description of bid
function sortbids(bids, count)
    local guilddump=eq.guildmembers()
    local alldkp=getalldkp()

    -- Determine if any main character has bid on the auction above the alt alt_bid_cap
    -- so that we know we have to cap alts
    local has_any_main_bid=false
    for bidder, bid in pairs(bids) do
        if not isalt(bidder, guilddump) and bid>alt_bid_cap then
            has_any_main_bid=true
            break
        end
    end

    -- Compute effective bids and display information
    local ordered_bids={}
    for bidder, bid in pairs(bids) do
        local dkp=getdkp(bidder, alldkp, guilddump)
        if dkp==nil then dkp="???" end
        local bid_entry={
            bidder=bidder,
            bid=bid,
            dkp=dkp
        }
        if isalt(bidder, guilddump) then
            bid_entry.display_name=initial_cap(bidder).."["..initial_cap(getmain(bidder, guilddump)).."]"
            if has_any_main_bid then
                bid_entry.effective_bid = math.min(alt_bid_cap, bid)
            else
                bid_entry.effective_bid = bid
            end
        else
            bid_entry.display_name=initial_cap(bidder)
            bid_entry.effective_bid = bid
        end
        table.insert(ordered_bids, bid_entry)
    end

    if #ordered_bids==0 then
        return 0, {}, {}
    end

    -- Sort by bids descending, then by character name alphabetically.
    table.sort(ordered_bids, function(l,r)
        return l.effective_bid>r.effective_bid or
                (l.effective_bid==r.effective_bid and l.bidder < r.bidder)
        end)

    -- Price is one more than the first loser (no more than the highest bid),
    -- or the minimum bid if there aren't enough bids
    local price
    if #ordered_bids<count+1 then
        price=minimum_bid
    else
        price=math.min(ordered_bids[1].effective_bid, 1+ordered_bids[count+1].effective_bid)
    end

    -- Collect results into the required return structures:  A list of winners
    -- and an ordered list of (winner name, bid info) pairs.
    local winners={}
    local displays={}
    for _, entry in ipairs(ordered_bids) do
        if entry.effective_bid>=price then
            table.insert(winners, entry.bidder)
        end
        if entry.bid == entry.effective_bid then
            table.insert(displays, {entry.display_name, entry.bid.."/"..entry.dkp})
        else
            table.insert(displays, {entry.display_name, entry.effective_bid.."("..entry.bid..")/"..entry.dkp})
        end
    end
    return price, winners, displays
end