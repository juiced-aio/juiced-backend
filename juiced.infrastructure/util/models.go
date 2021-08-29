package util

type CipherTextTooShortError struct{}

func (e *CipherTextTooShortError) Error() string {
	return "cipher text is too short"
}

type CipherTextNotMultipleOfBlockSizeError struct{}

func (e *CipherTextNotMultipleOfBlockSizeError) Error() string {
	return "cipher text not a multiple of the block size"
}

type CancellationToken struct {
	Cancel bool
}
