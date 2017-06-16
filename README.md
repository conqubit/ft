Can compute the discrete fourier transform of an image as an image, and can apply the inverse DFT to the resultant image and get the original image back. Currently only works on images with power-of-two dimensions. Additionally, amplitudes in the spectrum image only encode one component (brightness), so images will come back black and white if the inverse DFT is performed.

Usage:

Generate spectrum image of an image:
`imagedft <file>`

Generate original signal image from a spectrum image:
`imagedft -i <file>`
