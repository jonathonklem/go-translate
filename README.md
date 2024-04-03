# Go Translate
Simple program to use go routines to concurrently translate a text file to another language via sets of paragraphs. 

## Usage
./go-translate <filepath> <language> <routines>
filepath = path of file to translate
language = language to translate to
routines = # of concurrent routines to run

## Considerations
I have ChatGPT-4-Turbo hardcoded.  There are other options such as ChatGPT-4 and ChatGPT-3.5-Turbo.  I chose this as it seemed to provide the best translation and the most generous rate limits.

Additionally, there is a lot of pausing done in the routines.  This is intentionally.  While working with larger texts, the limiting factor always seemed to be ChatGPT's rate limits.  Specifically the # of tokens per minute that could be consumed.  