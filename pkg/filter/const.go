package filter

import "regexp"

var (
	cleanRE        = regexp.MustCompile(`\b(?:filter|sort)\[[^\]]+\]=[^&]+`)
	bracketRE      = regexp.MustCompile(`(?P<type>filter|sort|page)\[([^&]+?)\](\={1})`)
	bracketValueRE = regexp.MustCompile(`\]\=(.*?)(\&|\z)`)
	commaRE        = regexp.MustCompile(`\s?\,\s?`) //
	operatorRE     = regexp.MustCompile(`\[(?P<operator>nin|in|lt|lte|gt|gte|eq|ne)\]:`)
	valueRE        = regexp.MustCompile(`\]:(.*?)(\,\[|\z)`)
	fieldsRE       = regexp.MustCompile(`fields=(?P<field>.+?)(\&|\z)`)
	fieldsWithRE   = regexp.MustCompile(`fields\[([^&]+?)\]=([^&]+?)(\&|\z)`)
	groupByRE      = regexp.MustCompile(`group_by=(?P<field>.+?)(\&|\z)`)
	searchRE       = regexp.MustCompile(`search=(?P<field>[^&].+?)(\&|\z)`)
	numberRE       = regexp.MustCompile(`(\d*\.\d*$)`)
	objectIdRE     = regexp.MustCompile(`[0-9a-z]{24}$`)
	datetimeRE     = regexp.MustCompile(`(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})$`)
	timeRE         = regexp.MustCompile(`(\d{2}:\d{2}:\d{2})$`)
	dateRE         = regexp.MustCompile(`(\d{4}-\d{2}-\d{2})$`)
	uuidRE         = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)
