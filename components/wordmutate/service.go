/* Емкость мутатора.
 *
 * Берем максимальное растояние мутаций от оригинала 2 редактирования.
 * Если мы ограничим максимальную длину слова 24 рунами, то мутаций на удаление будет
 * не более 24*24 (+1, добавляем само слово) = 577.
 * А вот мутаций по вставке для русского алфавита (33 руны + дефис = 34)
 * получится уже 25 * 34 + 25 * 34 * 26 * 34 = 752250.
 * Это много, но терпимо. Но если мы не введем этих ограничений, потребление памяти и процессора
 * может оказаться неприемлемым. Согласно же частоте распределения русских слов по длине,
 * очень мало слов длинее 24 рун.
 * А в английском языке еще меньше.
 */

package wordmutate

type Component struct {
	ruAlphabet []rune
	enAlphabet []rune
}

func New() *Component {
	return &Component{
		ruAlphabet: []rune(`абвгдеёжзийклмнопрстуфхцчшщъыьэюя-`),
		enAlphabet: []rune("abcdefghijklmnopqrstuvwxyz-`'"),
	}
}

func (s *Component) Deletes(w string) []string {
	runeWord := []rune(w)
	if len(runeWord) == 1 || len(runeWord) > 24 {
		return nil
	}

	if len(runeWord) == 2 {
		return []string{
			w,
			string(runeWord[0]),
			string(runeWord[1]),
		}
	}

	res := make([]string, 0, len(runeWord)*len(runeWord)+1)
	res = append(res, w)

	deletedOne := deleteRune(w)
	res = append(res, deletedOne...)

	for _, v := range deletedOne {
		deleted := deleteRune(v)
		res = append(res, deleted...)
	}

	return res
}

func deleteRune(w string) []string {
	runeWord := []rune(w)
	res := make([]string, 0, len(runeWord))
	for i := 0; i < len(runeWord); i++ {
		res = append(res, string(runeWord[:i])+string(runeWord[i+1:]))
	}

	return res
}

func (s *Component) InsertRuneEn(w string) []string {
	runeWord := []rune(w)

	resLen := (len(runeWord)+1)*len(s.ruAlphabet) + 1

	res := make([]string, 0, resLen)
	for i := 0; i <= len(runeWord); i++ {
		for _, r := range s.enAlphabet {
			res = append(res, string(runeWord[:i])+string(r)+string(runeWord[i:]))
		}
	}

	return res
}

func (s *Component) InsertRuneRu(w string) []string {
	runeWord := []rune(w)

	resLen := (len(runeWord) + 1) * len(s.ruAlphabet)

	res := make([]string, 0, resLen)
	for i := 0; i <= len(runeWord); i++ {
		for _, r := range s.ruAlphabet {
			res = append(res, string(runeWord[:i])+string(r)+string(runeWord[i:]))
		}
	}

	return res
}
