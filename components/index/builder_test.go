package index

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cannonflesh/wordspell/components/langdetect"
	"github.com/cannonflesh/wordspell/testdata"
)

const (
	inHTML = "<h2 class=\"h4\"> Отряд щенков к делу готов!</h2><p>Колготки российского производства выполнены из нату-раль-ного и экологически чистого хлопка с небольшим процентом полиамида и эластана. Пусть вас не пугает наличие синтетических материалов, ведь благодаря им бельё: </p><ul><li> удобнее сидит на ножках; </li><li> не сползает; </li><li> лучше тянется; </li><li> легче надевается; </li><li> дольше служит. </li></ul><h2 class=\"h4\">Can`t</h2><ul><li> Оригинальный рисунок. </li><li> Приятная на ощупь ткань. </li><li> Изготовлено из отборной хлопковой пряжи наивысшего качества. </li></ul><p><b>Рекомендации по уходу</b>: стирка в бережном режиме при 40 °С. Вертикальная сушка. Осторожное глажение при температуре не более 110 °C. </p>"
	inText = "Шина \"N\" нулевая TDM, 6х9x200 мм, 4/1, 4 группы/крепеж по центру, SQ0801-0036"
)

func TestComponent_htmlPreProcess(t *testing.T) {
	check := []string{
		"отряд", "щенков", "делу", "готов",
		"колготки", "российского", "производства", "выполнены", "из", "нату-раль-ного", "экологически",
		"чистого", "хлопка", "небольшим", "процентом", "полиамида", "эластана",
		"пусть", "вас", "не", "пугает", "наличие", "синтетических", "материалов",
		"ведь", "благодаря", "им", "бельё", "удобнее", "сидит", "на", "ножках",
		"не", "сползает", "лучше", "тянется", "легче", "надевается", "дольше", "служит",
		"can`t", "оригинальный", "рисунок", "приятная", "на", "ощупь", "ткань", "изготовлено", "из",
		"отборной", "хлопковой", "пряжи", "наивысшего", "качества",
		"рекомендации", "по", "уходу",
		"стирка", "бережном", "режиме", "при",
		"вертикальная", "сушка", "осторожное", "глажение", "при", "температуре", "не", "более",
	}

	require.Equal(t, check, htmlPreProcess(inHTML))
}

func TestComponent_textPreProcess(t *testing.T) {
	require.Equal(t,
		[]string{"шина", "нулевая", "tdm", "мм", "группы", "крепеж", "по", "центру", "sq"},
		textPreProcess(inText),
	)
}

func TestComponent_processWordSlice(t *testing.T) {
	l, _ := testdata.NewTestLogger()
	langer := langdetect.New()

	source := NewMockDataSource(t)
	store := NewMockDataStore(t)

	d := NewBuilder(source, store, langer, l)
	ws := []string{"один", "два", "one", "три", "two", "четыре", "пять", "three", "four", "oneодин", "шесть"}
	dt := newData()
	d.processWordSlice(dt, ws)
	require.Len(t, dt.words[enLangCode], 4)
	require.Equal(t, frequency(1), dt.words[enLangCode]["four"])
	require.Len(t, dt.words[ruLangCode], 6)
	require.Equal(t, frequency(1), dt.words[ruLangCode]["пять"])
	require.Equal(t, frequency(1), dt.words[ruLangCode]["шесть"])
	require.Len(t, dt.dwords[enLangCode], 1)
	require.Equal(t, frequency(1), dt.dwords[enLangCode]["three four"])
	require.Len(t, dt.dwords[ruLangCode], 2)
	require.Equal(t, frequency(1), dt.dwords[ruLangCode]["четыре пять"])
	require.Equal(t, frequency(0), dt.words[ruLangCode]["oneодин"])
	require.Equal(t, frequency(0), dt.words[enLangCode]["oneодин"])
}
