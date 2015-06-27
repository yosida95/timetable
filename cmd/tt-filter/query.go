package main

type Cond struct {
	Title string `json:"title"`

	URL      string `json:"url"`
	MailAddr string `json:"mailaddr"`

	IsFirst  bool `json:"isFirst"`
	IsLive   bool `json:"isLive"`
	IsRepeat bool `json:"isRepeat"`
}

func (c *Cond) Filter() filter {
	f := make(AndFilter, 0, 6)

	if c.Title != "" {
		f = append(f, titleFilter(c.Title))
	}
	if c.URL != "" {
		f = append(f, urlFilter(c.URL))
	}
	if c.MailAddr != "" {
		f = append(f, mailAddrFilter(c.MailAddr))
	}
	if c.IsFirst {
		f = append(f, filterFunc(isFirstFilter))
	}
	if c.IsLive {
		f = append(f, filterFunc(isLiveFilter))
	}
	if c.IsRepeat {
		f = append(f, filterFunc(isRepeatFilter))
	}

	return f
}

type Query []Cond

func (q Query) Filter() filter {
	f := make(OrFilter, len(q))
	for i, c := range q {
		f[i] = c.Filter()
	}
	return f
}
